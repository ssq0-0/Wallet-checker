// Package wasmClient provides WebAssembly integration for cryptographic operations.
package wasmClient

import (
	"bytes"
	"chief-checker/pkg/logger"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WasmInstance represents a single WebAssembly instance with its runtime and functions.
// Each instance is isolated and can be used independently for cryptographic operations.
type WasmInstance struct {
	id            int64           // unique instance identifier
	ctx           context.Context // context for runtime operations
	runtime       wazero.Runtime  // WASM runtime
	module        api.Module      // loaded WASM module
	memory        api.Memory      // WASM memory
	fnMalloc      api.Function    // memory allocation function
	fnFree        api.Function    // memory deallocation function
	fnGetSignType api.Function    // signature type getter
	fnSetSignType api.Function    // signature type setter
	fnRS          api.Function    // random string generator
	fnMkSF        api.Function    // signature maker
}

// WasmClient manages a pool of WebAssembly instances for concurrent cryptographic operations.
// It provides automatic scaling and instance lifecycle management.
type WasmClient struct {
	WasmInstance *WasmInstance      // current instance
	mu           sync.Mutex         // protects instance pool
	instances    chan *WasmInstance // pool of available instances
	poolSize     int                // maximum pool size
	nextID       int64              // next instance ID
	activeCount  int32              // number of active instances
	totalCreated int32              // total instances created
	initialized  bool               // initialization flag
}

// NewWasm creates a new WasmClient with a pool of instances.
// It initializes the minimum required number of instances and prepares them for use.
//
// Returns:
// - Wasm: interface for cryptographic operations
// - error: if initialization fails
func NewWasm() (Wasm, error) {
	wasmService := &WasmClient{
		poolSize:  poolSize,
		instances: make(chan *WasmInstance, instanceCount),
	}

	if err := wasmService.Initialize(); err != nil {
		logger.GlobalLogger.Debugf("[WASM] Failed to initialize pool: %v", err)
		return nil, err
	}
	return wasmService, nil
}

// Initialize prepares the WasmClient for use by creating the initial pool of instances.
// This method is thread-safe and idempotent - it can be called multiple times safely.
//
// Returns an error if instance creation fails.
func (s *WasmClient) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		logger.GlobalLogger.Debugf("[WASM] Pool already initialized")
		return nil
	}

	logger.GlobalLogger.Debugf("[WASM] Initializing pool with size %d", s.poolSize)

	for i := 0; i < s.poolSize; i++ {
		instance, err := s.createInstance()
		if err != nil {
			return fmt.Errorf("failed to create instance %d: %v", i, err)
		}
		s.instances <- instance
	}

	s.initialized = true
	logger.GlobalLogger.Debugf("[WASM] Pool initialized with %d instances", s.poolSize)
	return nil
}

// createInstance creates a new WebAssembly instance with all necessary functions and memory.
// It initializes the WASM runtime, loads the module, and prepares all required functions.
//
// Returns:
// - *WasmInstance: newly created instance
// - error: if creation or initialization fails
func (s *WasmClient) createInstance() (*WasmInstance, error) {
	id := atomic.AddInt64(&s.nextID, 1)
	logger.GlobalLogger.Debugf("[WASM] Creating new instance #%d", id)

	instance := &WasmInstance{
		id:  id,
		ctx: context.Background(),
	}

	instance.runtime = wazero.NewRuntimeWithConfig(instance.ctx, wazero.NewRuntimeConfigInterpreter())

	wasi_snapshot_preview1.MustInstantiate(instance.ctx, instance.runtime)

	_, err := instance.runtime.NewHostModuleBuilder("env").
		NewFunctionBuilder().WithFunc(func() int32 { return 0 }).Export("iR").
		NewFunctionBuilder().WithFunc(func() int32 { return 1 }).Export("hNHD").
		Instantiate(instance.ctx)
	if err != nil {
		return nil, fmt.Errorf("env module failed: %v", err)
	}

	wasmBytes, err := base64.StdEncoding.DecodeString(pureWasmBase64)
	if err != nil {
		return nil, fmt.Errorf("decode wasm failed: %v", err)
	}

	compiled, err := instance.runtime.CompileModule(instance.ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("compile wasm failed: %v", err)
	}

	instance.module, err = instance.runtime.InstantiateModule(instance.ctx, compiled, wazero.NewModuleConfig())
	if err != nil {
		return nil, fmt.Errorf("instantiate wasm failed: %v", err)
	}

	instance.memory = instance.module.Memory()

	if _, ok := instance.memory.Grow(10); !ok {
		return nil, fmt.Errorf("memory grow failed")
	}

	instance.fnMalloc = instance.module.ExportedFunction("malloc")
	instance.fnFree = instance.module.ExportedFunction("free")
	instance.fnGetSignType = instance.module.ExportedFunction("get_sign_type")
	instance.fnSetSignType = instance.module.ExportedFunction("set_sign_type")
	instance.fnRS = instance.module.ExportedFunction("r_s")
	instance.fnMkSF = instance.module.ExportedFunction("mk_s_f")

	if _, err := instance.fnSetSignType.Call(instance.ctx, uint64(signTypeRegular)); err != nil {
		return nil, fmt.Errorf("set_sign_type failed: %v", err)
	}

	seed := time.Now().UnixNano()
	if !instance.memory.Write(seedOffset, binary.LittleEndian.AppendUint64(nil, uint64(seed))) {
		return nil, fmt.Errorf("failed to initialize random seed")
	}

	atomic.AddInt32(&s.totalCreated, 1)
	atomic.AddInt32(&s.activeCount, 1)
	logger.GlobalLogger.Debugf("[WASM] Instance #%d created (active: %d, total: %d)", id, atomic.LoadInt32(&s.activeCount), atomic.LoadInt32(&s.totalCreated))

	return instance, nil
}

// getInstance retrieves an instance from the pool or creates a new one if the pool is empty.
// This method is thread-safe and handles automatic instance creation when needed.
//
// Returns:
// - *WasmInstance: a ready-to-use WASM instance
// - error: if instance retrieval or creation fails
func (s *WasmClient) getInstance() (*WasmInstance, error) {
	select {
	case instance := <-s.instances:
		logger.GlobalLogger.Debugf("[WASM] Got instance #%d from pool (active: %d)", instance.id, atomic.LoadInt32(&s.activeCount))
		return instance, nil
	default:
		instance, err := s.createInstance()
		if err != nil {
			return nil, fmt.Errorf("failed to create new instance: %v", err)
		}
		return instance, nil
	}
}

// releaseInstance returns an instance to the pool or destroys it if the pool is full.
// This method is thread-safe and handles proper cleanup of resources.
//
// Parameters:
// - instance: the WASM instance to release
func (s *WasmClient) releaseInstance(instance *WasmInstance) {
	select {
	case s.instances <- instance:
		logger.GlobalLogger.Debugf("[WASM] Released instance #%d back to pool", instance.id)
	default:
		logger.GlobalLogger.Debugf("[WASM] Pool full, closing instance #%d", instance.id)
		instance.runtime.Close(instance.ctx)
		atomic.AddInt32(&s.activeCount, -1)
		logger.GlobalLogger.Debugf("[WASM] Instance #%d closed (active: %d)", instance.id, atomic.LoadInt32(&s.activeCount))
	}
}

// Close shuts down the WasmClient and releases all resources.
// This method is thread-safe and should be called when the client is no longer needed.
// After closing, the client cannot be reused.
//
// Returns an error if any instance fails to close properly.
func (s *WasmClient) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.GlobalLogger.Debugf("[WASM] Closing all instances")
	close(s.instances)
	for instance := range s.instances {
		if err := instance.runtime.Close(instance.ctx); err != nil {
			return fmt.Errorf("close runtime failed: %v", err)
		}
		atomic.AddInt32(&s.activeCount, -1)
		logger.GlobalLogger.Debugf("[WASM] Instance #%d closed (active: %d)", instance.id, atomic.LoadInt32(&s.activeCount))
	}

	return nil
}

// GenerateNonce generates a unique nonce using WASM cryptographic functions.
// The nonce is prefixed with either "n_" or "nc_" depending on the signature type.
//
// Returns:
// - string: generated nonce
// - error: if nonce generation fails
func (s *WasmClient) GenerateNonce() (string, error) {
	instance, err := s.getInstance()
	if err != nil {
		return "", fmt.Errorf("failed to get instance: %v", err)
	}
	defer s.releaseInstance(instance)

	seed := time.Now().UnixNano()
	if !instance.memory.Write(seedOffset, binary.LittleEndian.AppendUint64(nil, uint64(seed))) {
		return "", fmt.Errorf("failed to update random seed")
	}

	logger.GlobalLogger.Debugf("[WASM] Generating nonce with instance #%d", instance.id)

	size := uint32(40)
	noncePtr, err := s.malloc(instance, size)
	if err != nil {
		return "", fmt.Errorf("malloc nonce failed: %v", err)
	}
	defer s.free(instance, noncePtr)

	if _, err = instance.fnRS.Call(instance.ctx, uint64(size), uint64(noncePtr)); err != nil {
		return "", fmt.Errorf("rs call failed: %v", err)
	}

	raw, ok := instance.memory.Read(noncePtr, size)
	if !ok {
		return "", fmt.Errorf("read nonce failed: [%d:%d]", noncePtr, noncePtr+size)
	}

	typePtr, err := s.malloc(instance, 4)
	if err != nil {
		return "", fmt.Errorf("malloc type ptr failed: %v", err)
	}
	defer s.free(instance, typePtr)

	if _, err = instance.fnGetSignType.Call(instance.ctx, uint64(typePtr)); err != nil {
		return "", fmt.Errorf("get_sign_type failed: %v", err)
	}

	tBytes, ok := instance.memory.Read(typePtr, 4)
	if !ok {
		return "", fmt.Errorf("read type failed: [%d:%d]", typePtr, typePtr+4)
	}

	t := int32(binary.LittleEndian.Uint32(tBytes))
	prefix := "n_"
	if t == signTypeSecondary {
		prefix = "nc_"
	}

	nonce := prefix + string(bytes.TrimRight(raw, "\x00"))
	logger.GlobalLogger.Debugf("[WASM] Generated nonce with instance #%d: %s", instance.id, nonce)
	return nonce, nil
}

// MakeSignature creates a cryptographic signature for API requests.
// It combines multiple parameters to create a unique signature using WASM functions.
//
// Parameters:
// - method: HTTP method (GET, POST, etc.)
// - urlPath: API endpoint path
// - queryString: URL query parameters
// - nonce: unique request identifier
// - tsStr: timestamp string
//
// Returns:
// - string: generated signature
// - error: if signature generation fails
func (s *WasmClient) MakeSignature(method, urlPath, queryString, nonce, tsStr string) (string, error) {
	instance, err := s.getInstance()
	if err != nil {
		return "", fmt.Errorf("failed to get instance: %v", err)
	}
	defer s.releaseInstance(instance)

	logger.GlobalLogger.Debugf("[WASM] Making signature with instance #%d", instance.id)

	type memBlock struct {
		ptr  uint32
		data []byte
	}
	blocks := []memBlock{
		{data: append([]byte(method), 0)},
		{data: append([]byte(urlPath), 0)},
		{data: append([]byte(queryString), 0)},
		{data: append([]byte(nonce), 0)},
		{data: append([]byte(tsStr), 0)},
	}

	for i := range blocks {
		ptr, err := s.malloc(instance, uint32(len(blocks[i].data)))
		if err != nil {
			return "", fmt.Errorf("malloc failed: %v", err)
		}
		blocks[i].ptr = ptr
		if !instance.memory.Write(ptr, blocks[i].data) {
			return "", fmt.Errorf("write memory failed")
		}
	}
	defer func() {
		for _, b := range blocks {
			s.free(instance, b.ptr)
		}
	}()

	outPtr, err := s.malloc(instance, 512)
	if err != nil {
		return "", fmt.Errorf("malloc output failed: %v", err)
	}
	defer s.free(instance, outPtr)

	res, err := instance.fnMkSF.Call(instance.ctx,
		uint64(blocks[0].ptr),
		uint64(blocks[1].ptr),
		uint64(blocks[2].ptr),
		uint64(blocks[3].ptr),
		uint64(blocks[4].ptr),
		uint64(outPtr),
	)
	if err != nil {
		return "", fmt.Errorf("mk_s_f failed: %v", err)
	}

	sigSize := uint32(res[0])
	if sigSize == 0 {
		return "", fmt.Errorf("mk_s_f returned zero size")
	}

	sigBytes, ok := instance.memory.Read(outPtr, sigSize)
	if !ok {
		return "", fmt.Errorf("read signature failed: [%d:%d]", outPtr, outPtr+sigSize)
	}

	signature := string(bytes.TrimRight(sigBytes, "\x00"))
	logger.GlobalLogger.Debugf("[WASM] Generated signature with instance #%d: %s", instance.id, signature)
	return signature, nil
}

// malloc allocates memory in the WASM instance.
// This is an internal helper function for managing WASM memory.
//
// Parameters:
// - instance: WASM instance to allocate memory in
// - size: number of bytes to allocate
//
// Returns:
// - uint32: pointer to allocated memory
// - error: if allocation fails
func (s *WasmClient) malloc(instance *WasmInstance, size uint32) (uint32, error) {
	res, err := instance.fnMalloc.Call(instance.ctx, uint64(size))
	if err != nil {
		return 0, fmt.Errorf("malloc failed: %v", err)
	}

	ptr := uint32(res[0])
	if ptr == ^uint32(0) {
		return 0, fmt.Errorf("malloc returned invalid pointer")
	}

	logger.GlobalLogger.Debugf("[WASM] Allocated %d bytes at 0x%x in instance #%d", size, ptr, instance.id)
	return ptr, nil
}

// free deallocates memory in the WASM instance.
// This is an internal helper function for managing WASM memory.
//
// Parameters:
// - instance: WASM instance to free memory in
// - ptr: pointer to memory to free
//
// Returns an error if deallocation fails.
func (s *WasmClient) free(instance *WasmInstance, ptr uint32) error {
	_, err := instance.fnFree.Call(instance.ctx, uint64(ptr))
	if err != nil {
		return fmt.Errorf("free failed: %v", err)
	}
	logger.GlobalLogger.Debugf("[WASM] Freed memory at 0x%x in instance #%d", ptr, instance.id)
	return nil
}
