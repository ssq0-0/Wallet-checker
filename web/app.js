let amountsHidden = false;
let expandedAddresses = new Set();
let lastAddressesData = null;

function formatNumber(num) {
    if (amountsHidden) {
        return '****';
    }
    return new Intl.NumberFormat('ru-RU', {
        style: 'currency',
        currency: 'USD',
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
    }).format(num);
}

function formatPercentage(value, total) {
    if (total === 0) return '0.00%';
    return ((value / total) * 100).toFixed(2) + '%';
}

function formatAddress(address) {
    return `${address.slice(0, 6)}...${address.slice(-4)}`;
}

let topTokensChart = null;
let chainsChart = null;

let currentSort = {
    field: null,
    direction: 'desc'
};

function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

function updateChartData(chart, newData, duration = 500) {
    if (!chart || !newData) return;

    const oldData = chart.data.datasets[0].data;
    const newValues = newData.map(item => item.value || item.totalValue);
    
    chart.data.labels = newData.map(item => {
        if (item.symbol) {
            return item.symbol;
        } else if (item.name) {
            return `${item.name} (${formatNumber(item.totalValue)})`;
        }
        return '';
    });

    oldData.forEach((oldValue, index) => {
        const newValue = newValues[index];
        if (oldValue !== newValue) {
            animateValue(oldValue, newValue, duration, (value) => {
                chart.data.datasets[0].data[index] = value;
                chart.update('none');
            });
        }
    });

    if (newValues.length > oldData.length) {
        for (let i = oldData.length; i < newValues.length; i++) {
            chart.data.datasets[0].data[i] = newValues[i];
        }
        chart.update('none');
    }
}

function animateValue(start, end, duration, callback) {
    const startTime = performance.now();
    const change = end - start;

    function update(currentTime) {
        const elapsed = currentTime - startTime;
        const progress = Math.min(elapsed / duration, 1);
        
        const easeProgress = progress < 0.5
            ? 2 * progress * progress
            : 1 - Math.pow(-2 * progress + 2, 2) / 2;

        const currentValue = start + change * easeProgress;
        callback(currentValue);

        if (progress < 1) {
            requestAnimationFrame(update);
        }
    }

    requestAnimationFrame(update);
}

function createTopTokensChart(tokens) {
    console.log('Creating top tokens chart with data:', tokens);
    if (!tokens || tokens.length === 0) {
        console.error('No tokens data provided for chart');
        return;
    }

    const isDark = document.documentElement.classList.contains('dark');
    const textColor = isDark ? '#f3f4f6' : '#1f2937';
    const gridColor = isDark ? '#374151' : '#e5e7eb';

    const ctx = document.getElementById('topTokensChart').getContext('2d');
    
    if (topTokensChart) {
        updateChartData(topTokensChart, tokens);
        return;
    }
    
    topTokensChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: tokens.map(token => token.symbol),
            datasets: [{
                label: 'Стоимость в USD',
                data: tokens.map(token => token.value),
                backgroundColor: 'rgba(59, 130, 246, 0.5)',
                borderColor: 'rgba(59, 130, 246, 1)',
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: {
                duration: 0
            },
            plugins: {
                legend: {
                    display: false
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const value = context.raw;
                            return amountsHidden ? '****' : formatNumber(value);
                        }
                    }
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        callback: value => amountsHidden ? '****' : formatNumber(value),
                        color: textColor
                    },
                    grid: {
                        color: gridColor
                    }
                },
                x: {
                    ticks: {
                        color: textColor
                    },
                    grid: {
                        color: gridColor
                    }
                }
            }
        }
    });
}

function createChainsChart(chains) {
    console.log('Creating chains chart with data:', chains);
    if (!chains || chains.length === 0) {
        console.error('No chains data provided for chart');
        return;
    }

    const isDark = document.documentElement.classList.contains('dark');
    const textColor = isDark ? '#f3f4f6' : '#1f2937';

    const ctx = document.getElementById('chainsChart').getContext('2d');
    
    const sortedChains = [...chains]
        .sort((a, b) => b.totalValue - a.totalValue)
        .slice(0, 10);
    
    const colors = generateColors(sortedChains.length);
    const backgroundColors = colors.map(color => color + '80');
    const borderColors = colors;

    if (chainsChart) {
        chainsChart.data.labels = sortedChains.map(chain => 
            `${chain.name} (${amountsHidden ? '****' : formatNumber(chain.totalValue)})`
        );
        chainsChart.data.datasets[0].data = sortedChains.map(chain => chain.totalValue);
        chainsChart.data.datasets[0].backgroundColor = backgroundColors;
        chainsChart.data.datasets[0].borderColor = borderColors;
        chainsChart.update('none');
        return;
    }
    
    chainsChart = new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: sortedChains.map(chain => 
                `${chain.name} (${amountsHidden ? '****' : formatNumber(chain.totalValue)})`
            ),
            datasets: [{
                data: sortedChains.map(chain => chain.totalValue),
                backgroundColor: backgroundColors,
                borderColor: borderColors,
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: {
                duration: 0
            },
            plugins: {
                legend: {
                    position: 'right',
                    labels: {
                        color: textColor
                    }
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const value = context.raw;
                            return amountsHidden ? '****' : formatNumber(value);
                        }
                    }
                }
            }
        }
    });
}

function sortTable(field) {
    console.log('Sorting by field:', field);
    const header = document.querySelector(`th[data-sort="${field}"]`);
    
    document.querySelectorAll('th[data-sort]').forEach(th => {
        if (th !== header) {
            th.classList.remove('asc', 'desc');
        }
    });

    if (currentSort.field === field) {
        currentSort.direction = currentSort.direction === 'asc' ? 'desc' : 'asc';
    } else {
        currentSort.field = field;
        currentSort.direction = 'desc';
    }

    header.classList.remove('asc', 'desc');
    header.classList.add(currentSort.direction);

    if (lastAddressesData) {
        const sortedData = [...lastAddressesData].sort((a, b) => {
            let valueA, valueB;
            
            switch (field) {
                case 'totalBalance':
                    valueA = parseFloat(a.totalBalance) || 0;
                    valueB = parseFloat(b.totalBalance) || 0;
                    break;
                case 'tokenCount':
                    valueA = parseInt(a.tokenCount) || 0;
                    valueB = parseInt(b.tokenCount) || 0;
                    break;
                case 'projectCount':
                    valueA = parseInt(a.projectCount) || 0;
                    valueB = parseInt(b.projectCount) || 0;
                    break;
                default:
                    return 0;
            }

            if (currentSort.direction === 'asc') {
                return valueA - valueB;
            } else {
                return valueB - valueA;
            }
        });

        const tbody = document.getElementById('addressesTable');
        tbody.innerHTML = '';
        
        updateAddressesTable(sortedData);
    }
}

function updateAddressesTable(addresses) {
    if (!addresses || addresses.length === 0) return;
    
    lastAddressesData = addresses;

    const tbody = document.getElementById('addressesTable');
    const existingRows = new Map();
    
    tbody.querySelectorAll('tr[data-address]').forEach(row => {
        existingRows.set(row.dataset.address, row);
    });

    const existingAddresses = new Set(existingRows.keys());

    const newAddresses = addresses.filter(addr => !existingAddresses.has(addr.address));

    newAddresses.forEach(newAddr => {
        let insertAfterRow = null;
        
        if (currentSort.field) {
            const insertIndex = addresses.findIndex(addr => addr.address === newAddr.address);
            if (insertIndex > 0) {
                const prevAddr = addresses[insertIndex - 1];
                insertAfterRow = existingRows.get(prevAddr.address);
            }
        }

        const row = document.createElement('tr');
        row.dataset.address = newAddr.address;
        
        row.innerHTML = `
            <td class="px-4 py-2">
                <button class="expand-btn" onclick="toggleAddressExpansion('${newAddr.address}')">
                    ${expandedAddresses.has(newAddr.address) ? '▼' : '▶'}
                </button>
                <a href="https://debank.com/profile/${newAddr.address}" 
                   target="_blank" 
                   class="address-link">
                    ${formatAddress(newAddr.address)}
                </a>
            </td>
            <td class="px-4 py-2 text-right">${formatNumber(newAddr.totalBalance)}</td>
            <td class="px-4 py-2 text-right">${newAddr.tokenCount}</td>
            <td class="px-4 py-2 text-right">${newAddr.projectCount}</td>
        `;

        if (insertAfterRow) {
            insertAfterRow.after(row);
        } else {
            tbody.insertBefore(row, tbody.firstChild);
        }

        if (expandedAddresses.has(newAddr.address)) {
            const detailsRow = document.createElement('tr');
            detailsRow.classList.add('details-row');
            detailsRow.innerHTML = `
                <td colspan="4">
                    <div class="details-content">
                        <div class="tokens-section">
                            <h4>Топ токены</h4>
                            <div class="token-list">
                                ${newAddr.topTokens.map(token => `
                                    <div class="token-item">
                                        <span class="token-symbol">${token.symbol}</span>
                                        <span class="token-value">${formatNumber(token.value)}</span>
                                    </div>
                                `).join('')}
                            </div>
                        </div>
                        <div class="projects-section">
                            <h4>Топ проекты</h4>
                            <div class="project-list">
                                ${newAddr.topProjects.map(project => `
                                    <div class="project-item">
                                        <span class="project-name">${project.name}</span>
                                        <span class="project-value">${formatNumber(project.value)}</span>
                                    </div>
                                `).join('')}
                            </div>
                        </div>
                    </div>
                </td>
            `;
            row.after(detailsRow);
        }

        existingRows.set(newAddr.address, row);
    });

    addresses.forEach(addr => {
        const existingRow = existingRows.get(addr.address);
        if (existingRow) {
            const cells = existingRow.querySelectorAll('td');
            const totalBalanceCell = cells[1];
            const tokenCountCell = cells[2];
            const projectCountCell = cells[3];

            const newTotalBalance = formatNumber(addr.totalBalance);
            const newTokenCount = addr.tokenCount;
            const newProjectCount = addr.projectCount;

            if (totalBalanceCell.textContent !== newTotalBalance) {
                totalBalanceCell.textContent = newTotalBalance;
            }
            if (tokenCountCell.textContent !== newTokenCount) {
                tokenCountCell.textContent = newTokenCount;
            }
            if (projectCountCell.textContent !== newProjectCount) {
                projectCountCell.textContent = newProjectCount;
            }

            if (expandedAddresses.has(addr.address)) {
                const detailsRow = existingRow.nextElementSibling;
                if (detailsRow && detailsRow.classList.contains('details-row')) {
                    const tokenList = detailsRow.querySelector('.token-list');
                    const projectList = detailsRow.querySelector('.project-list');

                    const tokenItems = tokenList.querySelectorAll('.token-item');
                    addr.topTokens.forEach((token, index) => {
                        if (tokenItems[index]) {
                            const valueCell = tokenItems[index].querySelector('.token-value');
                            const newValue = formatNumber(token.value);
                            if (valueCell.textContent !== newValue) {
                                valueCell.textContent = newValue;
                            }
                        }
                    });

                    const projectItems = projectList.querySelectorAll('.project-item');
                    addr.topProjects.forEach((project, index) => {
                        if (projectItems[index]) {
                            const valueCell = projectItems[index].querySelector('.project-value');
                            const newValue = formatNumber(project.value);
                            if (valueCell.textContent !== newValue) {
                                valueCell.textContent = newValue;
                            }
                        }
                    });
                }
            }
        }
    });
}

function toggleAddressExpansion(address) {
    if (!lastAddressesData) return;
    
    const addressData = lastAddressesData.find(addr => addr.address === address);
    if (!addressData) return;

    const row = document.querySelector(`tr[data-address="${address}"]`);
    if (!row) return;

    const detailsRow = row.nextElementSibling;
    if (expandedAddresses.has(address)) {
        if (detailsRow && detailsRow.classList.contains('details-row')) {
            detailsRow.remove();
        }
        expandedAddresses.delete(address);
        row.querySelector('.expand-btn').textContent = '▶';
    } else {
        const newDetailsRow = document.createElement('tr');
        newDetailsRow.classList.add('details-row');
        newDetailsRow.innerHTML = `
            <td colspan="4">
                <div class="details-content">
                    <div class="tokens-section">
                        <h4>Топ токены</h4>
                        <div class="token-list">
                            ${addressData.topTokens.map(token => `
                                <div class="token-item">
                                    <span class="token-symbol">${token.symbol}</span>
                                    <span class="token-value">${formatNumber(token.value)}</span>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                    <div class="projects-section">
                        <h4>Топ проекты</h4>
                        <div class="project-list">
                            ${addressData.topProjects.map(project => `
                                <div class="project-item">
                                    <span class="project-name">${project.name}</span>
                                    <span class="project-value">${formatNumber(project.value)}</span>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                </div>
            </td>
        `;
        row.after(newDetailsRow);
        expandedAddresses.add(address);
        row.querySelector('.expand-btn').textContent = '▼';
    }
}

function updateGlobalStats(stats) {
    console.log('Updating global stats with data:', stats);
    if (!stats) {
        console.error('No stats data provided');
        return;
    }
    document.getElementById('totalAccounts').textContent = stats.totalAccounts;
    document.getElementById('totalValue').textContent = formatNumber(stats.totalUSDValue || 0);
}

function toggleAmounts() {
    amountsHidden = !amountsHidden;
    const button = document.getElementById('hideAmounts');
    button.textContent = amountsHidden ? 'Показать суммы' : 'Скрыть суммы';
    
    requestAnimationFrame(() => {
        if (chainsChart) {
            const newLabels = chainsChart.data.labels.map((label, index) => {
                const value = chainsChart.data.datasets[0].data[index];
                return `${label.split(' (')[0]} (${amountsHidden ? '****' : formatNumber(value)})`;
            });
            chainsChart.data.labels = newLabels;
            chainsChart.update('none');
        }

        if (topTokensChart) {
            topTokensChart.options.scales.y.ticks.callback = value => amountsHidden ? '****' : formatNumber(value);
            topTokensChart.update('none');
        }
        
        const totalValue = document.getElementById('totalValue');
        if (totalValue && lastBalanceData && lastBalanceData.globalStats) {
            totalValue.textContent = formatNumber(lastBalanceData.globalStats.totalUSDValue || 0);
        }

        if (lastAddressesData) {
            setTimeout(() => {
                updateAddressesTable(lastAddressesData);
            }, 50);
        }
    });
}

function hasDataChanged(oldData, newData) {
    if (!oldData || !newData) return true;
    
    if (oldData.globalStats.totalAccounts !== newData.globalStats.totalAccounts ||
        oldData.globalStats.totalUSDValue !== newData.globalStats.totalUSDValue) {
        return true;
    }
    
    if (JSON.stringify(oldData.topTokens) !== JSON.stringify(newData.topTokens)) {
        return true;
    }
    
    if (JSON.stringify(oldData.chains) !== JSON.stringify(newData.chains)) {
        return true;
    }
    
    return false;
}

let lastBalanceData = null;

async function loadData() {
    try {
        console.log('Fetching balance data...');
        const balanceResponse = await fetch('/api/balance');
        console.log('Balance response status:', balanceResponse.status);
        if (!balanceResponse.ok) {
            throw new Error(`HTTP error! status: ${balanceResponse.status}`);
        }
        const balanceData = await balanceResponse.json();
        console.log('Received balance data:', JSON.stringify(balanceData, null, 2));

        if (!balanceData.globalStats) {
            console.error('No globalStats in balance data');
            return;
        }

        if (hasDataChanged(lastBalanceData, balanceData)) {
            updateGlobalStats(balanceData.globalStats);

            if (balanceData.topTokens && balanceData.topTokens.length > 0) {
                createTopTokensChart(balanceData.topTokens);
            }

            if (balanceData.chains && balanceData.chains.length > 0) {
                createChainsChart(balanceData.chains);
            }

            lastBalanceData = balanceData;
        }

        console.log('Fetching addresses data...');
        const addressesResponse = await fetch('/api/addresses');
        console.log('Addresses response status:', addressesResponse.status);
        if (!addressesResponse.ok) {
            throw new Error(`HTTP error! status: ${addressesResponse.status}`);
        }
        const addressesData = await addressesResponse.json();
        console.log('Received addresses data:', JSON.stringify(addressesData, null, 2));

        const currentOrder = Array.from(document.getElementById('addressesTable').querySelectorAll('tr[data-address]'))
            .map(row => row.dataset.address);

        const addressesMap = new Map(addressesData.map(addr => [addr.address, addr]));

        const orderedAddresses = currentOrder
            .map(address => addressesMap.get(address))
            .filter(Boolean);

        addressesData.forEach(addr => {
            if (!currentOrder.includes(addr.address)) {
                if (currentSort.field) {
                    const insertIndex = orderedAddresses.findIndex(existingAddr => {
                        let valueA, valueB;
                        
                        switch (currentSort.field) {
                            case 'totalBalance':
                                valueA = parseFloat(addr.totalBalance) || 0;
                                valueB = parseFloat(existingAddr.totalBalance) || 0;
                                break;
                            case 'tokenCount':
                                valueA = parseInt(addr.tokenCount) || 0;
                                valueB = parseInt(existingAddr.tokenCount) || 0;
                                break;
                            case 'projectCount':
                                valueA = parseInt(addr.projectCount) || 0;
                                valueB = parseInt(existingAddr.projectCount) || 0;
                                break;
                            default:
                                return false;
                        }

                        return currentSort.direction === 'asc' ? valueA < valueB : valueA > valueB;
                    });

                    if (insertIndex === -1) {
                        orderedAddresses.push(addr);
                    } else {
                        orderedAddresses.splice(insertIndex, 0, addr);
                    }
                } else {
                    orderedAddresses.push(addr);
                }
            }
        });

        updateAddressesTable(orderedAddresses);
    } catch (error) {
        console.error('Error loading data:', error);
    }
}

function updateChartColors() {
    const isDark = document.documentElement.classList.contains('dark');
    const textColor = isDark ? '#f3f4f6' : '#1f2937';
    const gridColor = isDark ? '#374151' : '#e5e7eb';

    if (topTokensChart) {
        topTokensChart.options.scales.y.ticks.color = textColor;
        topTokensChart.options.scales.x.ticks.color = textColor;
        topTokensChart.options.scales.y.grid.color = gridColor;
        topTokensChart.options.scales.x.grid.color = gridColor;
        topTokensChart.update('none');
    }

    if (chainsChart) {
        chainsChart.options.plugins.legend.labels.color = textColor;
        chainsChart.update('none');
    }
}

function toggleTheme() {
    const html = document.documentElement;
    const isDark = html.classList.contains('dark');
    
    if (isDark) {
        html.classList.remove('dark');
        localStorage.setItem('theme', 'light');
    } else {
        html.classList.add('dark');
        localStorage.setItem('theme', 'dark');
    }

    document.querySelector('.sun-icon').classList.toggle('hidden');
    document.querySelector('.moon-icon').classList.toggle('hidden');

    updateChartColors();
}

function initTheme() {
    const savedTheme = localStorage.getItem('theme') || 'light';
    const html = document.documentElement;
    
    if (savedTheme === 'dark') {
        html.classList.remove('light');
        html.classList.add('dark');
        document.querySelector('.sun-icon').classList.add('hidden');
        document.querySelector('.moon-icon').classList.remove('hidden');
    } else {
        html.classList.remove('dark');
        html.classList.add('light');
        document.querySelector('.sun-icon').classList.remove('hidden');
        document.querySelector('.moon-icon').classList.add('hidden');
    }

    updateChartColors();
}

document.getElementById('themeToggle').addEventListener('click', toggleTheme);

document.addEventListener('DOMContentLoaded', () => {
    console.log('Page loaded, starting data updates...');
    
    initTheme();
    
    document.querySelectorAll('th[data-sort]').forEach(th => {
        th.addEventListener('click', () => {
            sortTable(th.dataset.sort);
        });
    });
    
    loadData();
    startPeriodicUpdate();

    document.getElementById('hideAmounts').addEventListener('click', toggleAmounts);
});

async function stopServer() {
    try {
        const response = await fetch('/api/stop', {
            method: 'POST'
        });
        
        if (response.ok) {
            alert('Сервер остановлен. Страница будет закрыта.');
            window.close();
        } else {
            alert('Ошибка при остановке сервера');
        }
    } catch (error) {
        console.error('Error stopping server:', error);
        alert('Ошибка при остановке сервера');
    }
}

document.getElementById('stopServer').addEventListener('click', stopServer);

async function startPeriodicUpdate() {
    while (true) {
        try {
            await loadData();
            await new Promise(resolve => setTimeout(resolve, 1000));
        } catch (error) {
            console.error('Error updating data:', error);
            await new Promise(resolve => setTimeout(resolve, 1000));
        }
    }
}

function generateColors(count) {
    const baseColors = [
        '#3B82F6',
        '#10B981',
        '#F59E0B',
        '#EF4444',
        '#8B5CF6',
        '#EC4899',
        '#14B8A6',
        '#F97316',
        '#6366F1',
        '#84CC16',
        '#06B6D4',
        '#A855F7',
        '#F43F5E',
        '#22C55E',
        '#EAB308',
        '#78716C',
        '#0891B2',
        '#7C3AED',
        '#BE185D',
        '#4F46E5'
    ];

    if (count > baseColors.length) {
        const colors = [...baseColors];
        while (colors.length < count) {
            const hue = Math.floor(Math.random() * 360);
            const saturation = 70 + Math.floor(Math.random() * 30);
            const lightness = 45 + Math.floor(Math.random() * 10);
            colors.push(`hsl(${hue}, ${saturation}%, ${lightness}%)`);
        }
        return colors;
    }

    return baseColors.slice(0, count);
}

async function takeScreenshot() {
    try {
        const canvas = await html2canvas(document.querySelector('.container'), {
            scale: 2,
            useCORS: true,
            logging: false,
            backgroundColor: document.documentElement.classList.contains('dark') ? '#111827' : '#f3f4f6'
        });

        const link = document.createElement('a');
        link.download = `balance-checker-${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.png`;
        link.href = canvas.toDataURL('image/png');
        link.click();
    } catch (error) {
        console.error('Ошибка при создании скриншота:', error);
        alert('Произошла ошибка при создании скриншота');
    }
}

document.getElementById('takeScreenshot').addEventListener('click', takeScreenshot);

const style = document.createElement('style');
style.textContent = `
    .expand-btn {
        background: none;
        border: none;
        cursor: pointer;
        padding: 6px;
        border-radius: 50%;
        transition: all 0.2s ease;
        color: inherit;
        font-size: 12px;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        width: 32px;
        height: 32px;
        margin-right: 8px;
        box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
    }

    .expand-btn:hover {
        background-color: rgba(0, 0, 0, 0.08);
        transform: translateY(-1px);
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    }

    .dark .expand-btn:hover {
        background-color: rgba(255, 255, 255, 0.15);
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    }

    .expand-btn:active {
        transform: translateY(1px) scale(0.95);
        box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
    }

    th[data-sort] {
        cursor: pointer;
        position: relative;
        padding-right: 24px !important;
        transition: background-color 0.2s ease;
    }

    th[data-sort]:hover {
        background-color: rgba(0, 0, 0, 0.05);
    }

    .dark th[data-sort]:hover {
        background-color: rgba(255, 255, 255, 0.05);
    }

    th[data-sort]::after {
        content: '';
        position: absolute;
        right: 8px;
        top: 50%;
        transform: translateY(-50%);
        width: 0;
        height: 0;
        border-left: 5px solid transparent;
        border-right: 5px solid transparent;
        opacity: 0.4;
        transition: opacity 0.2s ease;
    }

    th[data-sort].asc::after {
        border-bottom: 5px solid currentColor;
    }

    th[data-sort].desc::after {
        border-top: 5px solid currentColor;
    }

    th[data-sort]:hover::after {
        opacity: 0.8;
    }

    .address-link {
        text-decoration: none;
        color: inherit;
        transition: all 0.2s ease;
        padding: 4px 8px;
        border-radius: 6px;
    }

    .address-link:hover {
        color: #3B82F6;
        background-color: rgba(59, 130, 246, 0.1);
    }

    .dark .address-link:hover {
        color: #60A5FA;
        background-color: rgba(96, 165, 250, 0.1);
    }
`;
document.head.appendChild(style);

const logoOverlay = document.createElement('div');
logoOverlay.id = 'logo-overlay';
logoOverlay.innerHTML = `
  <div class="logo-anim-container">
    <img src="sticker.webp" alt="chief.ssq logo" class="chief-logo" />
    <div class="logo-title">chief.ssq</div>
  </div>
`;
document.body.appendChild(logoOverlay);

const style2 = document.createElement('style');
style2.textContent = `
#logo-overlay {
  position: fixed;
  z-index: 9999;
  inset: 0;
  background: #f3f4f6;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: opacity 1s cubic-bezier(.4,0,.2,1);
  opacity: 1;
}
.dark #logo-overlay { background: #111827; }
.logo-anim-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  animation: logo-pop 2s cubic-bezier(.4,0,.2,1);
}
.chief-logo {
  width: 180px;
  height: 180px;
  border-radius: 50%;
  box-shadow: 0 4px 24px rgba(0,0,0,0.10);
  background: #fff;
  object-fit: cover;
  margin-bottom: 18px;
  border: 4px solid #229ED9;
  display: block;
}
.logo-title {
  font-size: 2rem;
  font-weight: 700;
  color: #229ED9;
  letter-spacing: 2px;
  text-shadow: 0 2px 8px rgba(34,158,217,0.08);
}
@keyframes logo-pop {
  0% { opacity: 0; transform: scale(0.7); }
  60% { opacity: 1; transform: scale(1.08); }
  100% { opacity: 1; transform: scale(1); }
}
#logo-overlay.hide {
  opacity: 0;
  pointer-events: none;
}
`;
document.head.appendChild(style2);

document.addEventListener('DOMContentLoaded', () => {
  setTimeout(() => {
    logoOverlay.classList.add('hide');
    setTimeout(() => logoOverlay.remove(), 1200);
  }, 2200);
}); 