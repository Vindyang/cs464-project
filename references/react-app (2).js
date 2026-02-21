import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Link, useNavigate } from 'react-router-dom';

const customStyles = {
  root: {
    '--bg-canvas': '#FFFFFF',
    '--bg-subtle': '#FAFAFA',
    '--text-main': '#111111',
    '--text-secondary': '#666666',
    '--text-tertiary': '#999999',
    '--accent-primary': '#004EEB',
    '--accent-primary-hover': '#003CC5',
    '--accent-secondary': '#FF8866',
    '--border-color': '#E5E5E5',
    '--grid-line': '#F5F5F5',
    '--font-stack': "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif",
    '--font-mono': "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
    '--radius-sm': '2px',
    '--radius-md': '4px',
    '--space-unit': '8px'
  },
  body: {
    backgroundColor: 'var(--bg-canvas)',
    color: 'var(--text-main)',
    fontFamily: 'var(--font-stack)',
    WebkitFontSmoothing: 'antialiased',
    lineHeight: '1.5',
    height: '100vh',
    display: 'flex',
    flexDirection: 'column',
    overflow: 'hidden'
  },
  gridBg: {
    position: 'absolute',
    top: 0,
    left: 0,
    width: '100%',
    height: '100%',
    backgroundImage: 'linear-gradient(var(--grid-line) 1px, transparent 1px), linear-gradient(90deg, var(--grid-line) 1px, transparent 1px)',
    backgroundSize: '40px 40px',
    zIndex: -1,
    pointerEvents: 'none'
  },
  header: {
    height: '64px',
    borderBottom: '1px solid var(--border-color)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '0 24px',
    background: 'var(--bg-canvas)',
    zIndex: 10
  },
  brand: {
    display: 'flex',
    alignItems: 'center',
    gap: '12px',
    fontWeight: 600,
    fontSize: '16px'
  },
  brandIcon: {
    width: '24px',
    height: '24px',
    border: '1.5px solid var(--text-main)',
    position: 'relative'
  },
  brandIconAfter: {
    content: '""',
    position: 'absolute',
    top: '50%',
    left: '50%',
    transform: 'translate(-50%, -50%)',
    width: '8px',
    height: '8px',
    background: 'var(--accent-primary)'
  },
  main: {
    flex: 1,
    padding: '24px',
    maxWidth: '1400px',
    margin: '0 auto',
    width: '100%',
    display: 'flex',
    flexDirection: 'column',
    gap: '24px',
    overflow: 'hidden'
  },
  pageHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'flex-end'
  },
  filtersBar: {
    display: 'flex',
    gap: '12px',
    background: 'var(--bg-canvas)',
    padding: '12px',
    border: '1px solid var(--border-color)',
    position: 'relative'
  },
  filterGroup: {
    display: 'flex',
    flexDirection: 'column',
    gap: '4px'
  },
  inputMinimal: {
    border: '1px solid var(--border-color)',
    height: '32px',
    padding: '0 8px',
    fontSize: '13px',
    fontFamily: 'var(--font-stack)',
    background: 'var(--bg-subtle)',
    outline: 'none'
  },
  nodesContainer: {
    flex: 1,
    overflowY: 'auto',
    background: 'var(--bg-canvas)',
    border: '1px solid var(--border-color)',
    position: 'relative'
  },
  nodesTable: {
    width: '100%',
    borderCollapse: 'collapse',
    textAlign: 'left'
  },
  tableHeader: {
    position: 'sticky',
    top: 0,
    background: 'var(--bg-subtle)',
    padding: '12px 20px',
    borderBottom: '1px solid var(--border-color)',
    zIndex: 5
  },
  tableCell: {
    padding: '16px 20px',
    borderBottom: '1px solid var(--grid-line)',
    fontSize: '13px',
    verticalAlign: 'middle'
  },
  statusPill: {
    display: 'inline-flex',
    alignItems: 'center',
    gap: '6px',
    fontSize: '11px',
    fontFamily: 'var(--font-mono)',
    padding: '2px 8px',
    background: 'var(--bg-subtle)',
    border: '1px solid var(--border-color)',
    borderRadius: '10px'
  },
  dot: {
    width: '6px',
    height: '6px',
    borderRadius: '50%'
  },
  bandwidthMiniChart: {
    width: '80px',
    height: '16px',
    display: 'flex',
    alignItems: 'flex-end',
    gap: '1px'
  },
  bar: {
    flex: 1,
    background: 'var(--accent-primary)',
    opacity: 0.3
  },
  barActive: {
    flex: 1,
    background: 'var(--accent-primary)',
    opacity: 0.8
  }
};

const BrandIcon = () => (
  <div style={customStyles.brandIcon}>
    <div style={customStyles.brandIconAfter}></div>
  </div>
);

const Header = () => {
  const navigate = useNavigate();

  return (
    <header style={customStyles.header}>
      <div style={customStyles.brand}>
        <BrandIcon />
        <span>ZERO-STORE</span>
      </div>
      <nav style={{ display: 'flex', alignItems: 'center' }}>
        <Link
          to="/"
          style={{
            color: 'var(--text-secondary)',
            textDecoration: 'none',
            fontSize: '14px',
            marginLeft: '24px',
            transition: 'color 0.2s'
          }}
        >
          Dashboard
        </Link>
        <Link
          to="/nodes"
          style={{
            color: 'var(--text-main)',
            textDecoration: 'none',
            fontSize: '14px',
            marginLeft: '24px',
            transition: 'color 0.2s'
          }}
        >
          Nodes
        </Link>
        <Link
          to="/settings"
          style={{
            color: 'var(--text-secondary)',
            textDecoration: 'none',
            fontSize: '14px',
            marginLeft: '24px',
            transition: 'color 0.2s'
          }}
        >
          Settings
        </Link>
        <Link
          to="/api"
          style={{
            color: 'var(--text-secondary)',
            textDecoration: 'none',
            fontSize: '14px',
            marginLeft: '24px',
            transition: 'color 0.2s'
          }}
        >
          API
        </Link>
      </nav>
      <div style={{ display: 'flex', gap: '12px' }}>
        <button
          onClick={() => navigate('/register-node')}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '0 20px',
            height: '40px',
            fontSize: '14px',
            fontWeight: 500,
            cursor: 'pointer',
            transition: 'all 0.2s',
            borderRadius: 'var(--radius-sm)',
            backgroundColor: 'var(--accent-primary)',
            color: 'white',
            border: 'none'
          }}
        >
          Register Node
        </button>
      </div>
    </header>
  );
};

const CardInnerMarks = () => (
  <div style={{ position: 'absolute', width: '100%', height: '100%', pointerEvents: 'none' }}>
    <div style={{
      position: 'absolute',
      top: '-1px',
      left: '-1px',
      width: '6px',
      height: '6px',
      borderTop: '1px solid var(--text-tertiary)',
      borderLeft: '1px solid var(--text-tertiary)',
      opacity: 0.5
    }}></div>
    <div style={{
      position: 'absolute',
      top: '-1px',
      right: '-1px',
      width: '6px',
      height: '6px',
      borderTop: '1px solid var(--text-tertiary)',
      borderRight: '1px solid var(--text-tertiary)',
      opacity: 0.5
    }}></div>
  </div>
);

const CardBottomMarks = () => (
  <div style={{ position: 'absolute', width: '100%', height: '100%', pointerEvents: 'none' }}>
    <div style={{
      position: 'absolute',
      bottom: '-1px',
      left: '-1px',
      width: '6px',
      height: '6px',
      borderBottom: '1px solid var(--text-tertiary)',
      borderLeft: '1px solid var(--text-tertiary)',
      opacity: 0.5
    }}></div>
    <div style={{
      position: 'absolute',
      bottom: '-1px',
      right: '-1px',
      width: '6px',
      height: '6px',
      borderBottom: '1px solid var(--text-tertiary)',
      borderRight: '1px solid var(--text-tertiary)',
      opacity: 0.5
    }}></div>
  </div>
);

const BandwidthChart = ({ bars }) => (
  <div style={customStyles.bandwidthMiniChart}>
    {bars.map((bar, index) => (
      <div
        key={index}
        style={{
          ...(bar.active ? customStyles.barActive : customStyles.bar),
          height: bar.height
        }}
      ></div>
    ))}
  </div>
);

const StatusPill = ({ status, dotColor }) => (
  <div style={customStyles.statusPill}>
    <div style={{ ...customStyles.dot, background: dotColor }}></div>
    {status}
  </div>
);

const NodesPage = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState('All Nodes');
  const [regionFilter, setRegionFilter] = useState('Global');
  const [currentPage, setCurrentPage] = useState(2);

  const nodes = [
    {
      id: 'nd-7721-fx',
      status: 'ONLINE',
      dotColor: '#10B981',
      location: 'San Jose, US',
      ip: '192.168.1.42',
      uptime: '99.98%',
      bandwidth: '125 / 84 Mbps',
      capacity: '1.2 TB',
      capacityPercent: '85%',
      bars: [
        { height: '40%', active: false },
        { height: '60%', active: true },
        { height: '90%', active: true },
        { height: '70%', active: true },
        { height: '30%', active: false }
      ]
    },
    {
      id: 'nd-0922-zk',
      status: 'SYNCING',
      dotColor: '#004EEB',
      location: 'Frankfurt, DE',
      ip: '45.12.88.19',
      uptime: '94.12%',
      bandwidth: '412 / 210 Mbps',
      capacity: '512 GB',
      capacityPercent: '12%',
      bars: [
        { height: '20%', active: true },
        { height: '40%', active: true },
        { height: '30%', active: true },
        { height: '50%', active: true },
        { height: '80%', active: true }
      ]
    },
    {
      id: 'nd-4410-rt',
      status: 'ONLINE',
      dotColor: '#10B981',
      location: 'Tokyo, JP',
      ip: '102.14.5.21',
      uptime: '99.99%',
      bandwidth: '88 / 42 Mbps',
      capacity: '2.0 TB',
      capacityPercent: '44%',
      bars: [
        { height: '20%', active: false },
        { height: '20%', active: false },
        { height: '40%', active: true },
        { height: '30%', active: false },
        { height: '10%', active: false }
      ]
    },
    {
      id: 'nd-1182-pp',
      status: 'OFFLINE',
      dotColor: '#EF4444',
      location: 'London, UK',
      ip: '81.4.22.1',
      uptime: '82.44%',
      bandwidth: '0 / 0 Mbps',
      capacity: '1.0 TB',
      capacityPercent: '100%',
      bars: [
        { height: '5%', active: false },
        { height: '5%', active: false },
        { height: '5%', active: false },
        { height: '5%', active: false },
        { height: '5%', active: false }
      ]
    },
    {
      id: 'nd-6601-sq',
      status: 'ONLINE',
      dotColor: '#10B981',
      location: 'Sydney, AU',
      ip: '203.4.1.99',
      uptime: '99.95%',
      bandwidth: '55 / 22 Mbps',
      capacity: '4.0 TB',
      capacityPercent: '12%',
      bars: [
        { height: '50%', active: true },
        { height: '50%', active: true },
        { height: '60%', active: true },
        { height: '40%', active: true },
        { height: '30%', active: true }
      ]
    },
    {
      id: 'nd-3290-bc',
      status: 'ONLINE',
      dotColor: '#10B981',
      location: 'Singapore, SG',
      ip: '175.22.3.4',
      uptime: '99.90%',
      bandwidth: '210 / 190 Mbps',
      capacity: '2.5 TB',
      capacityPercent: '98%',
      bars: [
        { height: '80%', active: true },
        { height: '85%', active: true },
        { height: '90%', active: true },
        { height: '85%', active: true },
        { height: '80%', active: true }
      ]
    }
  ];

  return (
    <main style={customStyles.main}>
      <div style={customStyles.pageHeader}>
        <div>
          <h1 style={{ fontSize: '24px' }}>Network Nodes</h1>
          <p
            style={{
              fontFamily: 'var(--font-mono)',
              fontSize: '11px',
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
              color: 'var(--text-secondary)',
              marginTop: '4px'
            }}
          >
            142 Active Nodes / 2,408 Total Shards
          </p>
        </div>

        <div style={customStyles.filtersBar}>
          <CardInnerMarks />
          <div style={customStyles.filterGroup}>
            <span
              style={{
                fontFamily: 'var(--font-mono)',
                fontSize: '9px',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                color: 'var(--text-secondary)'
              }}
            >
              Search
            </span>
            <input
              type="text"
              style={{ ...customStyles.inputMinimal, width: '180px' }}
              placeholder="Node ID or IP..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
          <div style={customStyles.filterGroup}>
            <span
              style={{
                fontFamily: 'var(--font-mono)',
                fontSize: '9px',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                color: 'var(--text-secondary)'
              }}
            >
              Status
            </span>
            <select
              style={customStyles.inputMinimal}
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
            >
              <option>All Nodes</option>
              <option>Online</option>
              <option>Syncing</option>
              <option>Offline</option>
            </select>
          </div>
          <div style={customStyles.filterGroup}>
            <span
              style={{
                fontFamily: 'var(--font-mono)',
                fontSize: '9px',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                color: 'var(--text-secondary)'
              }}
            >
              Region
            </span>
            <select
              style={customStyles.inputMinimal}
              value={regionFilter}
              onChange={(e) => setRegionFilter(e.target.value)}
            >
              <option>Global</option>
              <option>North America</option>
              <option>Europe</option>
              <option>Asia Pacific</option>
            </select>
          </div>
        </div>
      </div>

      <div style={customStyles.nodesContainer}>
        <CardInnerMarks />
        <CardBottomMarks />
        <table style={customStyles.nodesTable}>
          <thead>
            <tr>
              <th
                style={{
                  ...customStyles.tableHeader,
                  fontFamily: 'var(--font-mono)',
                  fontSize: '11px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  color: 'var(--text-secondary)'
                }}
              >
                Node ID
              </th>
              <th
                style={{
                  ...customStyles.tableHeader,
                  fontFamily: 'var(--font-mono)',
                  fontSize: '11px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  color: 'var(--text-secondary)'
                }}
              >
                Status
              </th>
              <th
                style={{
                  ...customStyles.tableHeader,
                  fontFamily: 'var(--font-mono)',
                  fontSize: '11px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  color: 'var(--text-secondary)'
                }}
              >
                Geolocation
              </th>
              <th
                style={{
                  ...customStyles.tableHeader,
                  fontFamily: 'var(--font-mono)',
                  fontSize: '11px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  color: 'var(--text-secondary)'
                }}
              >
                Uptime
              </th>
              <th
                style={{
                  ...customStyles.tableHeader,
                  fontFamily: 'var(--font-mono)',
                  fontSize: '11px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  color: 'var(--text-secondary)'
                }}
              >
                Bandwidth (Up/Down)
              </th>
              <th
                style={{
                  ...customStyles.tableHeader,
                  fontFamily: 'var(--font-mono)',
                  fontSize: '11px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  color: 'var(--text-secondary)'
                }}
              >
                Capacity
              </th>
            </tr>
          </thead>
          <tbody>
            {nodes.map((node) => (
              <tr
                key={node.id}
                style={{ transition: 'background 0.2s' }}
                onMouseEnter={(e) => {
                  const cells = e.currentTarget.querySelectorAll('td');
                  cells.forEach((cell) => {
                    cell.style.background = 'var(--bg-subtle)';
                  });
                }}
                onMouseLeave={(e) => {
                  const cells = e.currentTarget.querySelectorAll('td');
                  cells.forEach((cell) => {
                    cell.style.background = '';
                  });
                }}
              >
                <td
                  style={{
                    ...customStyles.tableCell,
                    fontFamily: 'var(--font-mono)',
                    color: 'var(--text-main)',
                    fontWeight: 500
                  }}
                >
                  {node.id}
                </td>
                <td style={customStyles.tableCell}>
                  <StatusPill status={node.status} dotColor={node.dotColor} />
                </td>
                <td style={customStyles.tableCell}>
                  {node.location}{' '}
                  <span
                    style={{
                      color: 'var(--text-tertiary)',
                      fontFamily: 'var(--font-mono)',
                      fontSize: '11px'
                    }}
                  >
                    {node.ip}
                  </span>
                </td>
                <td style={{ ...customStyles.tableCell, fontFamily: 'var(--font-mono)' }}>
                  {node.uptime}
                </td>
                <td style={customStyles.tableCell}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                    <span style={{ fontFamily: 'var(--font-mono)' }}>{node.bandwidth}</span>
                    <BandwidthChart bars={node.bars} />
                  </div>
                </td>
                <td style={{ ...customStyles.tableCell, fontFamily: 'var(--font-mono)' }}>
                  {node.capacity}{' '}
                  <span style={{ color: 'var(--text-tertiary)' }}>({node.capacityPercent})</span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div style={{ display: 'flex', justifyContent: 'center', gap: '8px' }}>
        <button
          onClick={() => setCurrentPage(1)}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            height: '32px',
            width: '32px',
            padding: 0,
            fontSize: '14px',
            fontWeight: 500,
            cursor: 'pointer',
            transition: 'all 0.2s',
            borderRadius: 'var(--radius-sm)',
            background: 'transparent',
            border: '1px solid var(--border-color)',
            color: 'var(--text-main)'
          }}
        >
          1
        </button>
        <button
          onClick={() => setCurrentPage(2)}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            height: '32px',
            width: '32px',
            padding: 0,
            fontSize: '14px',
            fontWeight: 500,
            cursor: 'pointer',
            transition: 'all 0.2s',
            borderRadius: 'var(--radius-sm)',
            background: 'transparent',
            border: '1px solid var(--accent-primary)',
            color: 'var(--accent-primary)'
          }}
        >
          2
        </button>
        <button
          onClick={() => setCurrentPage(3)}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            height: '32px',
            width: '32px',
            padding: 0,
            fontSize: '14px',
            fontWeight: 500,
            cursor: 'pointer',
            transition: 'all 0.2s',
            borderRadius: 'var(--radius-sm)',
            background: 'transparent',
            border: '1px solid var(--border-color)',
            color: 'var(--text-main)'
          }}
        >
          3
        </button>
        <span
          style={{
            display: 'flex',
            alignItems: 'flex-end',
            padding: '0 4px',
            color: 'var(--text-tertiary)'
          }}
        >
          ...
        </span>
        <button
          onClick={() => setCurrentPage(24)}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            height: '32px',
            width: '32px',
            padding: 0,
            fontSize: '14px',
            fontWeight: 500,
            cursor: 'pointer',
            transition: 'all 0.2s',
            borderRadius: 'var(--radius-sm)',
            background: 'transparent',
            border: '1px solid var(--border-color)',
            color: 'var(--text-main)'
          }}
        >
          24
        </button>
      </div>
    </main>
  );
};

const App = () => {
  return (
    <Router basename="/">
      <div style={{ ...customStyles.root, ...customStyles.body }}>
        <div style={customStyles.gridBg}></div>
        <Header />
        <Routes>
          <Route path="/" element={<NodesPage />} />
          <Route path="/nodes" element={<NodesPage />} />
          <Route path="/settings" element={<NodesPage />} />
          <Route path="/api" element={<NodesPage />} />
          <Route path="/register-node" element={<NodesPage />} />
        </Routes>
      </div>
    </Router>
  );
};

export default App;