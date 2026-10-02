// SysMon Web Frontend Application - Observability Cockpit
(function () {
  'use strict';

  // State
  let currentView = 'overview';
  let isPaused = false;
  let eventSource = null;
  let pollInterval = null;
  let currentStats = null;
  let pendingKillPid = null;
  let pendingKillName = '';

  // Process table sort & filter
  let procSortCol = 'cpu_percent';
  let procSortAsc = false;
  let procSearchQuery = '';
  let procFilterType = 'all'; // 'all', 'high-cpu', 'high-mem', 'user'

  // DOM Elements
  const navItems = document.querySelectorAll('.nav-item');
  const viewPanels = document.querySelectorAll('.view-panel');
  const pageTitle = document.getElementById('page-title');
  const statusChip = document.getElementById('status-chip');
  const statusText = document.getElementById('status-text');
  const connIndicator = document.getElementById('conn-indicator');
  const btnPause = document.getElementById('btn-pause');
  const btnPauseText = document.getElementById('btn-pause-text');
  const btnPauseIconWrap = document.getElementById('btn-pause-icon-wrap');
  const refreshSelect = document.getElementById('refresh-select');
  const btnExport = document.getElementById('btn-export');
  const btnTheme = document.getElementById('btn-theme');
  const themeIconWrap = document.getElementById('theme-icon-wrap');
  const procSearchInput = document.getElementById('proc-search');
  const filterChips = document.querySelectorAll('.filter-chip');
  const killModal = document.getElementById('kill-modal');
  const killProcName = document.getElementById('kill-proc-name');
  const btnCancelKill = document.getElementById('btn-cancel-kill');
  const btnConfirmKill = document.getElementById('btn-confirm-kill');

  // Canvases
  const canvasCpu = document.getElementById('chart-cpu');
  const canvasMem = document.getElementById('chart-memory');
  const canvasCpuLarge = document.getElementById('chart-cpu-large');
  const canvasMemLarge = document.getElementById('chart-mem-large');
  const canvasNet = document.getElementById('chart-net');

  // Chart data history buffers (max 60 points)
  let cpuHistory = [];
  let memHistory = [];
  let netDownHistory = [];
  let netUpHistory = [];

  // SVG Icon Templates
  const ICONS = {
    pause: `<svg class="icon-svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="6" y="4" width="4" height="16"></rect><rect x="14" y="4" width="4" height="16"></rect></svg>`,
    play: `<svg class="icon-svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg>`,
    sun: `<svg class="icon-svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="5"></circle><line x1="12" y1="1" x2="12" y2="3"></line><line x1="12" y1="21" x2="12" y2="23"></line><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line><line x1="1" y1="12" x2="3" y2="12"></line><line x1="21" y1="12" x2="23" y2="12"></line><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line></svg>`,
    moon: `<svg class="icon-svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path></svg>`,
  };

  const viewTitles = {
    overview: 'System Overview',
    'cpu-mem': 'CPU & Memory Diagnostics',
    processes: 'Process Explorer',
    network: 'Network Traffic & Interfaces',
    disks: 'Storage Volumes & Disks',
    system: 'Host & System Information',
  };

  // Initialize
  function init() {
    initTheme();
    setupEventListeners();
    connectSSE();
  }

  // Theme Management
  function initTheme() {
    const savedTheme = localStorage.getItem('sysmon-theme') || 'dark';
    applyTheme(savedTheme);
  }

  function applyTheme(theme) {
    if (theme === 'light') {
      document.body.classList.remove('theme-dark');
      document.body.classList.add('theme-light');
      themeIconWrap.innerHTML = ICONS.moon;
      btnTheme.title = 'Switch to Dark Mode';
    } else {
      document.body.classList.remove('theme-light');
      document.body.classList.add('theme-dark');
      themeIconWrap.innerHTML = ICONS.sun;
      btnTheme.title = 'Switch to Light Mode';
    }
    localStorage.setItem('sysmon-theme', theme);
    redrawCharts();
  }

  function toggleTheme() {
    const isLight = document.body.classList.contains('theme-light');
    applyTheme(isLight ? 'dark' : 'light');
  }

  // Event Listeners
  function setupEventListeners() {
    // Navigation
    navItems.forEach((btn) => {
      btn.addEventListener('click', () => {
        const view = btn.dataset.view;
        switchView(view);
      });
    });

    // Theme toggle
    btnTheme.addEventListener('click', toggleTheme);

    // Pause toggle
    btnPause.addEventListener('click', togglePause);

    // Refresh rate selector
    refreshSelect.addEventListener('change', (e) => {
      const ms = parseInt(e.target.value, 10);
      updateSettings({ refresh_rate_ms: ms });
    });

    // Export button
    btnExport.addEventListener('click', exportData);

    // Process search input
    procSearchInput.addEventListener('input', (e) => {
      procSearchQuery = e.target.value.toLowerCase().trim();
      renderProcesses();
    });

    // Filter Chips
    filterChips.forEach((chip) => {
      chip.addEventListener('click', () => {
        filterChips.forEach((c) => c.classList.remove('active'));
        chip.classList.add('active');
        procFilterType = chip.dataset.filter;
        renderProcesses();
      });
    });

    // Process table sorting
    document.querySelectorAll('#procs-table th.sortable').forEach((th) => {
      th.addEventListener('click', () => {
        const col = th.dataset.sort;
        if (procSortCol === col) {
          procSortAsc = !procSortAsc;
        } else {
          procSortCol = col;
          procSortAsc = false;
        }
        updateSortHeaders();
        renderProcesses();
      });
    });

    // Kill Modal
    btnCancelKill.addEventListener('click', () => {
      killModal.classList.remove('open');
      pendingKillPid = null;
    });

    btnConfirmKill.addEventListener('click', () => {
      if (pendingKillPid) {
        killProcess(pendingKillPid);
      }
      killModal.classList.remove('open');
      pendingKillPid = null;
    });

    // Window resize handler with debounce
    let resizeTimer;
    window.addEventListener('resize', () => {
      clearTimeout(resizeTimer);
      resizeTimer = setTimeout(redrawCharts, 100);
    });

    // Ensure continuous updating even across tab visibility changes
    document.addEventListener('visibilitychange', () => {
      if (!document.hidden) {
        if (!eventSource || eventSource.readyState === EventSource.CLOSED) {
          connectSSE();
        } else {
          fetchStatsFallback();
        }
      }
    });
  }

  function switchView(view) {
    currentView = view;
    navItems.forEach((item) => {
      item.classList.toggle('active', item.dataset.view === view);
    });
    viewPanels.forEach((panel) => {
      panel.classList.toggle('active', panel.id === `view-${view}`);
    });
    pageTitle.textContent = viewTitles[view] || 'System Observability';
    redrawCharts();
  }

  // SSE & API Connection
  function connectSSE() {
    if (eventSource) {
      eventSource.close();
    }

    setConnectionStatus('connecting');

    eventSource = new EventSource('/api/events');

    eventSource.onopen = () => {
      setConnectionStatus(isPaused ? 'paused' : 'live');
      if (pollInterval) {
        clearInterval(pollInterval);
        pollInterval = null;
      }
    };

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        handleStatsUpdate(data);
      } catch (err) {
        console.error('Error parsing telemetry frame:', err);
      }
    };

    eventSource.onerror = () => {
      setConnectionStatus('error');
      eventSource.close();
      eventSource = null;

      // Fallback polling
      if (!pollInterval) {
        pollInterval = setInterval(fetchStatsFallback, 2500);
      }
      setTimeout(connectSSE, 5000);
    };
  }

  function fetchStatsFallback() {
    fetch('/api/stats')
      .then((res) => res.json())
      .then((data) => {
        setConnectionStatus(data.paused ? 'paused' : 'live');
        handleStatsUpdate(data);
      })
      .catch(() => {
        setConnectionStatus('error');
      });
  }

  function setConnectionStatus(status) {
    statusChip.className = 'status-chip ' + status;
    connIndicator.className = 'pulse-indicator ' + status;

    if (status === 'live') {
      statusText.textContent = 'LIVE';
    } else if (status === 'paused') {
      statusText.textContent = 'PAUSED';
    } else if (status === 'connecting') {
      statusText.textContent = 'CONNECTING';
    } else {
      statusText.textContent = 'OFFLINE';
    }
  }

  function togglePause() {
    isPaused = !isPaused;
    btnPauseText.textContent = isPaused ? 'Resume' : 'Pause';
    btnPauseIconWrap.innerHTML = isPaused ? ICONS.play : ICONS.pause;
    setConnectionStatus(isPaused ? 'paused' : 'live');
    updateSettings({ paused: isPaused });
  }

  function updateSettings(settings) {
    fetch('/api/settings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(settings),
    }).catch((err) => console.error('Settings update error:', err));
  }

  function exportData() {
    window.location.href = '/api/export';
  }

  function killProcess(pid) {
    fetch('/api/process/kill', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ pid: pid }),
    })
      .then((res) => res.json())
      .then((res) => {
        if (res.error) {
          alert('Failed to terminate process: ' + res.error);
        } else {
          fetchStatsFallback();
        }
      })
      .catch((err) => alert('Termination error: ' + err));
  }

  // Handle Incoming Data
  function handleStatsUpdate(data) {
    currentStats = data;
    if (data.paused !== undefined && isPaused !== data.paused) {
      isPaused = data.paused;
      btnPauseText.textContent = isPaused ? 'Resume' : 'Pause';
      btnPauseIconWrap.innerHTML = isPaused ? ICONS.play : ICONS.pause;
      setConnectionStatus(isPaused ? 'paused' : 'live');
    }

    // Update histories
    if (data.history) {
      cpuHistory = data.history.cpu || [];
      memHistory = data.history.memory || [];
      netDownHistory = (data.history.network || []).map((p) => p.download_kbps || 0);
      netUpHistory = (data.history.network || []).map((p) => p.upload_kbps || 0);
    }

    renderHeaderAndSidebar(data);
    renderOverview(data);
    renderCpuMem(data);
    renderProcesses(data);
    renderNetwork(data);
    renderDisks(data);
    renderSystem(data);
    redrawCharts();
  }

  const faviconLink = document.querySelector("link[rel*='icon']");
  function updateDynamicFavicon(cpu) {
    if (!faviconLink) return;
    const color = cpu > 80 ? '%23ef4444' : cpu > 60 ? '%23f59e0b' : '%230284c7';
    const svg = `%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'%3E%3Crect width='32' height='32' rx='8' fill='${color}'/%3E%3Cpath d='M6 16h5l3-7 4 14 3-7h5' stroke='white' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round' fill='none'/%3E%3C/svg%3E`;
    faviconLink.href = 'data:image/svg+xml,' + svg;
  }

  // Render Sidebar & Header
  function renderHeaderAndSidebar(data) {
    const host = data.system?.host;
    if (host) {
      document.getElementById('sidebar-hostname').textContent = host.hostname || 'Unknown Host';
      document.getElementById('sidebar-uptime').textContent = 'Up: ' + formatUptime(host.uptime);
    }

    const procs = data.processes;
    if (procs) {
      document.getElementById('badge-procs').textContent = procs.total_processes || 0;
    }

    // Live Taskbar & Tab Title Updates (continuously updating)
    const cpuVal = (data.system?.cpu?.usage || 0).toFixed(0);
    const memVal = (data.system?.memory?.used_percent || 0).toFixed(0);
    document.title = `⚡ CPU ${cpuVal}% · RAM ${memVal}% — SysMon`;
    updateDynamicFavicon(parseInt(cpuVal, 10));
  }

  // Render View 1: Overview
  function renderOverview(data) {
    const sys = data.system;
    if (!sys) return;

    // CPU Card
    const cpuUsage = (sys.cpu?.usage || 0).toFixed(1);
    document.getElementById('card-cpu-val').textContent = `${cpuUsage}%`;
    document.getElementById('bar-cpu').style.width = `${Math.min(100, Math.max(0, cpuUsage))}%`;
    document.getElementById('card-cpu-cores').textContent = `${sys.cpu?.cores || 0} Cores`;
    document.getElementById('card-cpu-model').textContent = sys.cpu?.model_name || 'Generic CPU';
    document.getElementById('chart-cpu-cur').textContent = `${cpuUsage}%`;

    // Memory Card
    const mem = sys.memory;
    if (mem) {
      const memUsage = (mem.used_percent || 0).toFixed(1);
      document.getElementById('card-mem-val').textContent = `${memUsage}%`;
      document.getElementById('bar-mem').style.width = `${Math.min(100, Math.max(0, memUsage))}%`;
      document.getElementById('card-mem-fraction').textContent = `${formatBytes(mem.used)} / ${formatBytes(mem.total)}`;
      document.getElementById('card-mem-avail').textContent = `Available: ${formatBytes(mem.available)}`;
      document.getElementById('chart-mem-cur').textContent = `${memUsage}%`;
    }

    // Disk Card
    if (sys.disk && sys.disk.length > 0) {
      const primaryDisk = sys.disk[0];
      const diskUsage = (primaryDisk.used_percent || 0).toFixed(1);
      document.getElementById('card-disk-val').textContent = `${diskUsage}%`;
      document.getElementById('bar-disk').style.width = `${Math.min(100, Math.max(0, diskUsage))}%`;
      document.getElementById('card-disk-fraction').textContent = `${formatBytes(primaryDisk.used)} / ${formatBytes(primaryDisk.total)}`;
      document.getElementById('card-disk-name').textContent = `${primaryDisk.mountpoint || primaryDisk.device} (${primaryDisk.fstype})`;
    }

    // Network Speeds
    let downKb = 0;
    let upKb = 0;
    if (data.speeds && data.speeds.length > 0) {
      data.speeds.forEach((s) => {
        downKb += s.download_kbps || 0;
        upKb += s.upload_kbps || 0;
      });
    }
    document.getElementById('card-net-down').textContent = formatSpeed(downKb);
    document.getElementById('card-net-up').textContent = formatSpeed(upKb);
    if (data.network) {
      document.getElementById('card-net-conns').textContent = `${data.network.connections || 0} Conns`;
      document.getElementById('card-net-total').textContent = `Total: ${formatBytes(data.network.total_recv)} Rx / ${formatBytes(data.network.total_sent)} Tx`;
    }

    // Top Workloads
    const procs = data.processes;
    if (procs) {
      // Top CPU
      const topCpuBody = document.getElementById('top-cpu-tbody');
      const topCpuList = procs.top_cpu ? procs.top_cpu.slice(0, 5) : [];
      topCpuBody.innerHTML = topCpuList
        .map(
          (p) => `
        <tr>
          <td>${p.pid}</td>
          <td><strong>${escapeHtml(p.name)}</strong></td>
          <td style="color: var(--accent-amber)">${p.cpu_percent.toFixed(1)}%</td>
          <td>${formatBytes(p.memory_mb * 1024 * 1024)}</td>
        </tr>`
        )
        .join('');

      // Top Memory
      const topMemBody = document.getElementById('top-mem-tbody');
      const topMemList = procs.top_memory ? procs.top_memory.slice(0, 5) : [];
      topMemBody.innerHTML = topMemList
        .map(
          (p) => `
        <tr>
          <td>${p.pid}</td>
          <td><strong>${escapeHtml(p.name)}</strong></td>
          <td>${formatBytes(p.memory_mb * 1024 * 1024)}</td>
          <td style="color: var(--accent-cyan)">${p.mem_percent.toFixed(1)}%</td>
        </tr>`
        )
        .join('');
    }
  }

  // Render View 2: CPU & Memory
  function renderCpuMem(data) {
    const mem = data.system?.memory;
    if (!mem) return;

    document.getElementById('mem-bd-total').textContent = formatBytes(mem.total);
    document.getElementById('mem-bd-used').textContent = `${formatBytes(mem.used)} (${mem.used_percent.toFixed(1)}%)`;
    document.getElementById('mem-bd-avail').textContent = formatBytes(mem.available);
    document.getElementById('mem-bd-free').textContent = formatBytes(mem.free);
    document.getElementById('mem-bd-cached').textContent = formatBytes(mem.cached);
    document.getElementById('mem-bd-buffers').textContent = formatBytes(mem.buffers);

    // Render Logical Processors Activity (Task Manager Style)
    const perCore = data.system?.cpu?.per_core;
    const coresGrid = document.getElementById('cpu-cores-grid');
    const coresSummary = document.getElementById('cores-summary');
    if (perCore && perCore.length > 0 && coresGrid) {
      if (coresSummary) coresSummary.textContent = `${perCore.length} Cores Active`;
      coresGrid.innerHTML = perCore
        .map((coreUsage, idx) => {
          const val = Math.min(100, Math.max(0, coreUsage)).toFixed(1);
          const color = coreUsage > 80 ? 'var(--accent-rose)' : coreUsage > 60 ? 'var(--accent-amber)' : 'var(--accent-cyan)';
          return `
          <div class="core-cell">
            <div class="core-cell-header">
              <span class="core-cell-num">Core ${idx}</span>
              <span class="core-cell-val" style="color: ${color}">${val}%</span>
            </div>
            <div class="core-bar-bg">
              <div class="core-bar-fill" style="width: ${val}%; background: ${color}"></div>
            </div>
          </div>`;
        })
        .join('');
    }
  }

  // Render View 3: Processes
  function renderProcesses(data) {
    const procs = data?.processes || currentStats?.processes;
    if (!procs) return;

    document.getElementById('procs-total-badge').textContent = `Total: ${procs.total_processes || 0}`;
    document.getElementById('procs-running-badge').textContent = `Running: ${procs.running_processes || 0}`;
    document.getElementById('procs-sleeping-badge').textContent = `Sleeping: ${procs.sleeping_processes || 0}`;

    let list = (procs.all_processes || []).slice();

    // Quick Filter Chips
    if (procFilterType === 'high-cpu') {
      list = list.filter((p) => p.cpu_percent > 5);
    } else if (procFilterType === 'high-mem') {
      list = list.filter((p) => p.memory_mb > 100);
    } else if (procFilterType === 'user') {
      list = list.filter((p) => p.username && !p.username.toLowerCase().includes('system'));
    }

    // Search Query Filter
    if (procSearchQuery) {
      list = list.filter((p) => {
        return (
          p.name.toLowerCase().includes(procSearchQuery) ||
          String(p.pid).includes(procSearchQuery) ||
          (p.username && p.username.toLowerCase().includes(procSearchQuery))
        );
      });
    }

    // Sort
    list.sort((a, b) => {
      let valA = a[procSortCol];
      let valB = b[procSortCol];

      if (typeof valA === 'string') {
        valA = valA.toLowerCase();
        valB = (valB || '').toLowerCase();
        return procSortAsc ? valA.localeCompare(valB) : valB.localeCompare(valA);
      }
      valA = valA || 0;
      valB = valB || 0;
      return procSortAsc ? valA - valB : valB - valA;
    });

    const tbody = document.getElementById('processes-tbody');
    if (list.length === 0) {
      tbody.innerHTML = '<tr><td colspan="8" class="text-center">No matching processes found</td></tr>';
      return;
    }

    // Render top 120 processes
    const renderLimit = 120;
    const itemsToRender = list.slice(0, renderLimit);

    tbody.innerHTML = itemsToRender
      .map((p) => {
        const memBytes = p.memory_mb * 1024 * 1024;
        const cpuColor = p.cpu_percent > 50 ? 'var(--accent-rose)' : p.cpu_percent > 20 ? 'var(--accent-amber)' : 'inherit';
        return `
        <tr>
          <td>${p.pid}</td>
          <td><strong>${escapeHtml(p.name)}</strong></td>
          <td style="color: var(--text-muted)">${escapeHtml(p.username || '-')}</td>
          <td style="color: ${cpuColor}">${p.cpu_percent.toFixed(1)}%</td>
          <td>${p.mem_percent.toFixed(1)}%</td>
          <td>${formatBytes(memBytes)}</td>
          <td><span class="badge ${p.status === 'running' ? 'badge-success' : 'badge-dim'}">${p.status || 'active'}</span></td>
          <td>
            <button class="btn btn-sm btn-danger" onclick="window.confirmKill(${p.pid}, '${escapeHtml(p.name)}')" aria-label="Kill process ${p.pid}">Kill</button>
          </td>
        </tr>`;
      })
      .join('');
  }

  window.confirmKill = function (pid, name) {
    pendingKillPid = pid;
    pendingKillName = name;
    killProcName.textContent = `${name} (PID: ${pid})`;
    killModal.classList.add('open');
  };

  function updateSortHeaders() {
    document.querySelectorAll('#procs-table th.sortable').forEach((th) => {
      const col = th.dataset.sort;
      const baseLabel = th.childNodes[0].textContent.trim();
      const sortArr = th.querySelector('.sort-arr');
      if (col === procSortCol) {
        th.className = `sortable active-sort ${procSortAsc ? 'asc' : 'desc'}`;
        if (sortArr) sortArr.textContent = procSortAsc ? '▲' : '▼';
      } else {
        th.className = 'sortable';
        if (sortArr) sortArr.textContent = '↕';
      }
    });
  }

  // Render View 4: Network
  function renderNetwork(data) {
    const net = data.network;
    if (!net || !net.interfaces) return;

    const container = document.getElementById('interfaces-container');
    container.innerHTML = net.interfaces
      .map((iface) => {
        const isUp = iface.is_up;
        return `
        <div class="card">
          <div class="card-header">
            <span class="card-title">${escapeHtml(iface.name)}</span>
            <span class="badge ${isUp ? 'badge-success' : 'badge-dim'}">${isUp ? 'ACTIVE' : 'INACTIVE'}</span>
          </div>
          <div class="meta-list">
            <div class="meta-item"><dt>Total Rx (Received)</dt><dd class="tabular">${formatBytes(iface.bytes_recv)}</dd></div>
            <div class="meta-item"><dt>Total Tx (Transmitted)</dt><dd class="tabular">${formatBytes(iface.bytes_sent)}</dd></div>
            <div class="meta-item"><dt>Packets (Rx / Tx)</dt><dd class="tabular">${iface.packets_recv.toLocaleString()} / ${iface.packets_sent.toLocaleString()}</dd></div>
            <div class="meta-item"><dt>Errors (In / Out)</dt><dd class="tabular">${iface.errin} / ${iface.errout}</dd></div>
            <div class="meta-item"><dt>Dropped (In / Out)</dt><dd class="tabular">${iface.dropin} / ${iface.dropout}</dd></div>
          </div>
        </div>`;
      })
      .join('');
  }

  // Render View 5: Disks
  function renderDisks(data) {
    const disks = data.system?.disk;
    if (!disks) return;

    const container = document.getElementById('disks-container');
    container.innerHTML = disks
      .map((d) => {
        const pct = (d.used_percent || 0).toFixed(1);
        const barColor = pct > 85 ? 'var(--accent-rose)' : pct > 70 ? 'var(--accent-amber)' : 'var(--accent-emerald)';
        return `
        <div class="card">
          <div class="card-header">
            <div>
              <span class="card-title">${escapeHtml(d.mountpoint || d.device)}</span>
              <div class="card-sub">${escapeHtml(d.device)} (${escapeHtml(d.fstype)})</div>
            </div>
            <span class="badge tabular" style="font-size: 13px; font-weight: 700;">${pct}%</span>
          </div>
          <div class="progress-bar-bg" style="height: 8px; margin-bottom: 14px;">
            <div class="progress-bar-fill" style="width: ${Math.min(100, Math.max(0, pct))}%; background: ${barColor};"></div>
          </div>
          <div class="meta-list">
            <div class="meta-item"><dt>Capacity</dt><dd class="tabular">${formatBytes(d.total)}</dd></div>
            <div class="meta-item"><dt>Used Space</dt><dd class="tabular">${formatBytes(d.used)}</dd></div>
            <div class="meta-item"><dt>Free Available</dt><dd class="tabular">${formatBytes(d.free)}</dd></div>
          </div>
        </div>`;
      })
      .join('');
  }

  // Render View 6: System
  function renderSystem(data) {
    const sys = data.system;
    if (!sys) return;

    const host = sys.host;
    if (host) {
      document.getElementById('sys-hostname').textContent = host.hostname || '-';
      document.getElementById('sys-os').textContent = host.os || '-';
      document.getElementById('sys-platform').textContent = host.platform || '-';
      document.getElementById('sys-kernel').textContent = host.kernel_version || '-';
      document.getElementById('sys-uptime').textContent = formatUptime(host.uptime);
    }

    if (sys.cpu) {
      document.getElementById('sys-cpu-model').textContent = sys.cpu.model_name || '-';
      document.getElementById('sys-cpu-cores').textContent = `${sys.cpu.cores} Logical Cores`;
    }

    if (sys.memory) {
      document.getElementById('sys-ram-total').textContent = formatBytes(sys.memory.total);
    }

    if (sys.disk) {
      document.getElementById('sys-disk-count').textContent = `${sys.disk.length} Disks`;
    }

    if (data.network?.interfaces) {
      document.getElementById('sys-net-count').textContent = `${data.network.interfaces.length} Interfaces`;
    }
  }

  // Chart Rendering
  function redrawCharts() {
    const isDark = document.body.classList.contains('theme-dark');
    const gridColor = isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)';
    const textColor = isDark ? '#64748b' : '#94a3b8';

    if (currentView === 'overview') {
      drawLineChart(canvasCpu, cpuHistory, {
        color: '#38bdf8',
        fillColor: 'rgba(56, 189, 248, 0.12)',
        max: 100,
        unit: '%',
        gridColor,
        textColor,
      });

      drawLineChart(canvasMem, memHistory, {
        color: '#a855f7',
        fillColor: 'rgba(168, 85, 247, 0.12)',
        max: 100,
        unit: '%',
        gridColor,
        textColor,
      });
    } else if (currentView === 'cpu-mem') {
      drawLineChart(canvasCpuLarge, cpuHistory, {
        color: '#38bdf8',
        fillColor: 'rgba(56, 189, 248, 0.15)',
        max: 100,
        unit: '%',
        gridColor,
        textColor,
        height: 230,
      });

      drawLineChart(canvasMemLarge, memHistory, {
        color: '#a855f7',
        fillColor: 'rgba(168, 85, 247, 0.15)',
        max: 100,
        unit: '%',
        gridColor,
        textColor,
        height: 230,
      });
    } else if (currentView === 'network') {
      drawDualLineChart(canvasNet, netDownHistory, netUpHistory, {
        color1: '#10b981',
        color2: '#38bdf8',
        unit: ' KB/s',
        gridColor,
        textColor,
      });
    }
  }

  function drawLineChart(canvas, points, options) {
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    const dpr = window.devicePixelRatio || 1;
    const rect = canvas.getBoundingClientRect();

    if (rect.width === 0) return;

    canvas.width = rect.width * dpr;
    canvas.height = (options.height || 180) * dpr;
    ctx.scale(dpr, dpr);

    const w = rect.width;
    const h = options.height || 180;
    const padding = { top: 18, right: 18, bottom: 24, left: 48 };

    ctx.clearRect(0, 0, w, h);

    const chartW = w - padding.left - padding.right;
    const chartH = h - padding.top - padding.bottom;
    const maxVal = options.max || Math.max(10, ...points, 1);

    // Guide Lines
    ctx.strokeStyle = options.gridColor;
    ctx.fillStyle = options.textColor;
    ctx.font = '10px ui-monospace, monospace';
    ctx.lineWidth = 1;

    for (let i = 0; i <= 4; i++) {
      const y = padding.top + (chartH / 4) * i;
      const val = Math.round(maxVal - (maxVal / 4) * i);
      ctx.beginPath();
      ctx.moveTo(padding.left, y);
      ctx.lineTo(w - padding.right, y);
      ctx.stroke();
      ctx.fillText(`${val}${options.unit || ''}`, 6, y + 3);
    }

    if (!points || points.length === 0) return;

    const stepX = chartW / Math.max(59, points.length - 1);
    const startX = padding.left + Math.max(60 - points.length, 0) * stepX;

    // Line Path
    ctx.beginPath();
    points.forEach((val, i) => {
      const x = startX + i * stepX;
      const y = padding.top + chartH - (Math.min(val, maxVal) / maxVal) * chartH;
      if (i === 0) ctx.moveTo(x, y);
      else ctx.lineTo(x, y);
    });

    ctx.strokeStyle = options.color;
    ctx.lineWidth = 2;
    ctx.lineJoin = 'round';
    ctx.stroke();

    // Area Fill
    if (points.length > 1) {
      const lastX = startX + (points.length - 1) * stepX;
      ctx.lineTo(lastX, padding.top + chartH);
      ctx.lineTo(startX, padding.top + chartH);
      ctx.closePath();
      ctx.fillStyle = options.fillColor;
      ctx.fill();

      // Pulsing Current Point Dot
      const currentVal = points[points.length - 1];
      const curX = lastX;
      const curY = padding.top + chartH - (Math.min(currentVal, maxVal) / maxVal) * chartH;

      ctx.beginPath();
      ctx.arc(curX, curY, 4, 0, Math.PI * 2);
      ctx.fillStyle = options.color;
      ctx.fill();
    }
  }

  function drawDualLineChart(canvas, points1, points2, options) {
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    const dpr = window.devicePixelRatio || 1;
    const rect = canvas.getBoundingClientRect();

    if (rect.width === 0) return;

    canvas.width = rect.width * dpr;
    canvas.height = 210 * dpr;
    ctx.scale(dpr, dpr);

    const w = rect.width;
    const h = 210;
    const padding = { top: 18, right: 18, bottom: 24, left: 56 };

    ctx.clearRect(0, 0, w, h);

    const chartW = w - padding.left - padding.right;
    const chartH = h - padding.top - padding.bottom;
    const maxVal = Math.max(10, ...points1, ...points2, 1);

    ctx.strokeStyle = options.gridColor;
    ctx.fillStyle = options.textColor;
    ctx.font = '10px ui-monospace, monospace';
    ctx.lineWidth = 1;

    for (let i = 0; i <= 4; i++) {
      const y = padding.top + (chartH / 4) * i;
      const val = Math.round(maxVal - (maxVal / 4) * i);
      ctx.beginPath();
      ctx.moveTo(padding.left, y);
      ctx.lineTo(w - padding.right, y);
      ctx.stroke();
      ctx.fillText(`${val}${options.unit || ''}`, 6, y + 3);
    }

    function renderSeries(points, color) {
      if (!points || points.length === 0) return;
      const stepX = chartW / Math.max(59, points.length - 1);
      const startX = padding.left + Math.max(60 - points.length, 0) * stepX;

      ctx.beginPath();
      points.forEach((val, i) => {
        const x = startX + i * stepX;
        const y = padding.top + chartH - (Math.min(val, maxVal) / maxVal) * chartH;
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
      });
      ctx.strokeStyle = color;
      ctx.lineWidth = 2;
      ctx.lineJoin = 'round';
      ctx.stroke();
    }

    renderSeries(points1, options.color1);
    renderSeries(points2, options.color2);
  }

  // Format Helpers
  function formatBytes(bytes) {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  function formatSpeed(kbps) {
    if (!kbps || kbps === 0) return '0.0 KB/s';
    if (kbps >= 1024) {
      return (kbps / 1024).toFixed(1) + ' MB/s';
    }
    return kbps.toFixed(1) + ' KB/s';
  }

  function formatUptime(seconds) {
    if (!seconds) return '0m';
    const d = Math.floor(seconds / (3600 * 24));
    const h = Math.floor((seconds % (3600 * 24)) / 3600);
    const m = Math.floor((seconds % 3600) / 60);

    if (d > 0) return `${d}d ${h}h ${m}m`;
    if (h > 0) return `${h}h ${m}m`;
    return `${m}m`;
  }

  function escapeHtml(str) {
    if (!str) return '';
    return String(str)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

  window.addEventListener('DOMContentLoaded', init);
})();
