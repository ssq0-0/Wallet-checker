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

type WasmInstance struct {
	id            int64
	ctx           context.Context
	runtime       wazero.Runtime
	module        api.Module
	memory        api.Memory
	fnMalloc      api.Function
	fnFree        api.Function
	fnGetSignType api.Function
	fnSetSignType api.Function
	fnRS          api.Function
	fnMkSF        api.Function
}

type WasmClient struct {
	WasmInstance *WasmInstance
	mu           sync.Mutex
	instances    chan *WasmInstance
	poolSize     int
	nextID       int64
	activeCount  int32
	totalCreated int32
	initialized  bool
}

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

// MakeSignature создает подпись через WASM
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

// malloc выделяет память в WASM
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

// free освобождает память в WASM
func (s *WasmClient) free(instance *WasmInstance, ptr uint32) error {
	_, err := instance.fnFree.Call(instance.ctx, uint64(ptr))
	if err != nil {
		return fmt.Errorf("free failed: %v", err)
	}
	logger.GlobalLogger.Debugf("[WASM] Freed memory at 0x%x in instance #%d", ptr, instance.id)
	return nil
}
