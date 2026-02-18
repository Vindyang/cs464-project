import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';

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
  }
};

const GridBackground = () => (
  <div style={{
    position: 'absolute',
    top: 0,
    left: 0,
    width: '100%',
    height: '100%',
    backgroundImage: 'linear-gradient(var(--grid-line) 1px, transparent 1px), linear-gradient(90deg, var(--grid-line) 1px, transparent 1px)',
    backgroundSize: '40px 40px',
    zIndex: -1,
    pointerEvents: 'none'
  }} />
);

const BrandIcon = () => (
  <div style={{
    width: '24px',
    height: '24px',
    border: '1.5px solid var(--text-main)',
    position: 'relative'
  }}>
    <div style={{
      position: 'absolute',
      top: '50%',
      left: '50%',
      transform: 'translate(-50%, -50%)',
      width: '8px',
      height: '8px',
      background: 'var(--accent-primary)'
    }} />
  </div>
);

const CardCorners = () => (
  <>
    <div style={{
      position: 'absolute',
      top: '-1px',
      left: '-1px',
      width: '6px',
      height: '6px',
      borderColor: 'var(--text-tertiary)',
      borderStyle: 'solid',
      borderWidth: '1px 0 0 1px',
      opacity: 0.5
    }} />
    <div style={{
      position: 'absolute',
      top: '-1px',
      right: '-1px',
      width: '6px',
      height: '6px',
      borderColor: 'var(--text-tertiary)',
      borderStyle: 'solid',
      borderWidth: '1px 1px 0 0',
      opacity: 0.5
    }} />
    <div style={{
      position: 'absolute',
      bottom: '-1px',
      left: '-1px',
      width: '6px',
      height: '6px',
      borderColor: 'var(--text-tertiary)',
      borderStyle: 'solid',
      borderWidth: '0 0 1px 1px',
      opacity: 0.5
    }} />
    <div style={{
      position: 'absolute',
      bottom: '-1px',
      right: '-1px',
      width: '6px',
      height: '6px',
      borderColor: 'var(--text-tertiary)',
      borderStyle: 'solid',
      borderWidth: '0 1px 1px 0',
      opacity: 0.5
    }} />
  </>
);

const Header = ({ activeTab, setActiveTab }) => (
  <header style={{
    height: '64px',
    borderBottom: '1px solid var(--border-color)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '0 24px',
    background: 'var(--bg-canvas)',
    zIndex: 10
  }}>
    <div style={{
      display: 'flex',
      alignItems: 'center',
      gap: '12px',
      fontWeight: 600,
      fontSize: '16px'
    }}>
      <BrandIcon />
      <span>ZERO-STORE</span>
    </div>
    <nav>
      {['Dashboard', 'Nodes', 'Settings', 'API'].map((tab) => (
        <a
          key={tab}
          href="#"
          onClick={(e) => {
            e.preventDefault();
            setActiveTab(tab);
          }}
          style={{
            color: activeTab === tab ? 'var(--text-main)' : 'var(--text-secondary)',
            textDecoration: 'none',
            fontSize: '14px',
            marginLeft: '24px',
            transition: 'color 0.2s'
          }}
        >
          {tab}
        </a>
      ))}
    </nav>
    <div style={{ display: 'flex', gap: '12px' }}>
      <a
        href="#"
        onClick={(e) => e.preventDefault()}
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
          textDecoration: 'none',
          backgroundColor: 'var(--accent-primary)',
          color: 'white',
          border: 'none'
        }}
      >
        Upload File
      </a>
    </div>
  </header>
);

const StorageOverview = () => (
  <section style={{
    gridColumn: '1 / 2',
    gridRow: '1 / 2',
    background: 'var(--bg-canvas)',
    border: '1px solid var(--border-color)',
    padding: '20px',
    position: 'relative',
    display: 'flex',
    flexDirection: 'column'
  }}>
    <CardCorners />
    <div style={{
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'baseline',
      marginBottom: '20px'
    }}>
      <span style={{
        fontFamily: 'var(--font-mono)',
        fontSize: '11px',
        textTransform: 'uppercase',
        letterSpacing: '0.05em',
        color: 'var(--text-secondary)'
      }}>Storage Used</span>
      <span style={{
        fontFamily: 'var(--font-mono)',
        fontSize: '11px',
        textTransform: 'uppercase',
        letterSpacing: '0.05em',
        color: 'var(--accent-primary)'
      }}>Healthy</span>
    </div>
    
    <div style={{
      fontSize: '48px',
      fontWeight: 600,
      letterSpacing: '-0.04em',
      lineHeight: 1
    }}>
      4.2<span style={{ fontSize: '20px', color: 'var(--text-tertiary)' }}>TB</span>
    </div>
    <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginTop: '4px' }}>
      of 12.0 TB Total Capacity
    </div>

    <div style={{ marginTop: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px' }}>
        <span>Usage Distribution</span>
        <span style={{ fontFamily: 'var(--font-mono)' }}>35%</span>
      </div>
      <div style={{
        height: '4px',
        background: 'var(--grid-line)',
        marginTop: '8px',
        position: 'relative',
        overflow: 'hidden'
      }}>
        <div style={{
          height: '100%',
          background: 'var(--accent-primary)',
          width: '35%',
          transition: 'width 1s ease-out'
        }} />
      </div>
    </div>

    <div style={{
      display: 'grid',
      gridTemplateColumns: '1fr 1fr',
      gap: '16px',
      marginTop: '24px',
      paddingTop: '24px',
      borderTop: '1px solid var(--grid-line)'
    }}>
      <div>
        <div style={{
          fontFamily: 'var(--font-mono)',
          fontSize: '11px',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
          color: 'var(--text-secondary)',
          marginBottom: '4px'
        }}>Objects</div>
        <div style={{
          fontSize: '24px',
          fontWeight: 500,
          letterSpacing: '-0.03em'
        }}>842k</div>
      </div>
      <div>
        <div style={{
          fontFamily: 'var(--font-mono)',
          fontSize: '11px',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
          color: 'var(--text-secondary)',
          marginBottom: '4px'
        }}>Providers</div>
        <div style={{
          fontSize: '24px',
          fontWeight: 500,
          letterSpacing: '-0.03em'
        }}>5</div>
      </div>
    </div>
  </section>
);

const ProviderMatrix = () => {
  const providers = [
    { icon: 'AWS', name: 'Amazon S3', region: 'us-east-1', latency: '24ms', usage: 78, status: 'green' },
    { icon: 'GCP', name: 'Google Drive', region: 'global', latency: '45ms', usage: 42, status: 'green' },
    { icon: 'DBX', name: 'Dropbox', region: 'eu-west', latency: '112ms', usage: 15, status: 'blue' },
    { icon: 'B2', name: 'Backblaze B2', region: 'us-west', latency: '89ms', usage: 91, status: 'green' },
    { icon: 'MS', name: 'OneDrive', region: 'us-east', latency: '320ms', usage: 5, status: 'orange' }
  ];

  const statusColors = {
    green: '#10B981',
    blue: 'var(--accent-primary)',
    orange: 'var(--accent-secondary)'
  };

  return (
    <div style={{
      gridColumn: '2 / 3',
      gridRow: '1 / 3',
      display: 'flex',
      flexDirection: 'column',
      gap: '24px'
    }}>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: '16px'
      }}>
        <h2 style={{ fontSize: '18px' }}>Storage Providers</h2>
        <a
          href="#"
          onClick={(e) => e.preventDefault()}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '0 20px',
            height: '32px',
            fontSize: '12px',
            fontWeight: 500,
            cursor: 'pointer',
            transition: 'all 0.2s',
            borderRadius: 'var(--radius-sm)',
            textDecoration: 'none',
            background: 'transparent',
            border: '1px solid var(--border-color)',
            color: 'var(--text-main)'
          }}
        >
          Manage Providers
        </a>
      </div>

      <div style={{
        background: 'var(--bg-canvas)',
        border: '1px solid var(--border-color)',
        padding: 0,
        position: 'relative'
      }}>
        <CardCorners />
        
        <div style={{
          display: 'grid',
          gridTemplateColumns: '40px 2fr 1fr 1fr 1fr',
          padding: '0 20px',
          alignItems: 'center',
          background: 'var(--bg-subtle)',
          height: '40px',
          fontSize: '11px',
          textTransform: 'uppercase',
          color: 'var(--text-secondary)',
          letterSpacing: '0.05em',
          borderBottom: '1px solid var(--border-color)'
        }}>
          <div></div>
          <div>Provider</div>
          <div>Region</div>
          <div>Latency</div>
          <div>Usage</div>
        </div>

        {providers.map((provider, index) => (
          <div
            key={index}
            style={{
              display: 'grid',
              gridTemplateColumns: '40px 2fr 1fr 1fr 1fr',
              padding: '16px 20px',
              alignItems: 'center',
              borderBottom: index < providers.length - 1 ? '1px solid var(--border-color)' : 'none',
              transition: 'background 0.1s'
            }}
          >
            <div style={{
              width: '24px',
              height: '24px',
              background: '#F0F0F0',
              borderRadius: '2px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: '10px',
              fontWeight: 'bold',
              color: 'var(--text-secondary)'
            }}>
              {provider.icon}
            </div>
            <div style={{ fontWeight: 500 }}>{provider.name}</div>
            <div style={{
              fontFamily: 'var(--font-mono)',
              fontSize: '12px',
              color: 'var(--text-secondary)'
            }}>
              {provider.region}
            </div>
            <div style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '6px',
              fontSize: '12px',
              color: 'var(--text-secondary)'
            }}>
              <div style={{
                width: '6px',
                height: '6px',
                borderRadius: '50%',
                background: statusColors[provider.status]
              }} />
              {provider.latency}
            </div>
            <div style={{
              height: '4px',
              background: '#eee',
              margin: 0,
              width: '60px',
              position: 'relative',
              overflow: 'hidden'
            }}>
              <div style={{
                height: '100%',
                background: 'var(--accent-primary)',
                width: `${provider.usage}%`
              }} />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

const SystemHealth = () => {
  const shards = [
    ...Array(9).fill('data'),
    ...Array(6).fill('parity')
  ];

  return (
    <section style={{
      gridColumn: '3 / 4',
      gridRow: '1 / 2',
      background: 'var(--bg-canvas)',
      border: '1px solid var(--border-color)',
      padding: '20px',
      position: 'relative',
      display: 'flex',
      flexDirection: 'column'
    }}>
      <CardCorners />
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'baseline',
        marginBottom: '20px'
      }}>
        <span style={{
          fontFamily: 'var(--font-mono)',
          fontSize: '11px',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
          color: 'var(--text-secondary)'
        }}>Shard Health (R-S 9,6)</span>
        <span style={{
          fontFamily: 'var(--font-mono)',
          fontSize: '11px',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
          color: 'var(--accent-primary)'
        }}>Optimal</span>
      </div>

      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(5, 1fr)',
        gap: '4px',
        margin: '20px 0'
      }}>
        {shards.map((type, index) => (
          <div
            key={index}
            style={{
              aspectRatio: '1',
              background: type === 'data' ? 'rgba(0, 78, 235, 0.05)' : 'rgba(255, 136, 102, 0.05)',
              border: type === 'data' ? '1px solid var(--accent-primary)' : '1px solid var(--accent-secondary)',
              position: 'relative'
            }}
            title={type === 'data' ? `Data Shard ${index + 1}` : `Parity Shard ${index - 8}`}
          >
            <div style={{
              position: 'absolute',
              top: '2px',
              right: '2px',
              width: '2px',
              height: '2px',
              background: 'currentColor',
              opacity: 0.5
            }} />
          </div>
        ))}
      </div>
      
      <div style={{
        fontSize: '12px',
        color: 'var(--text-secondary)',
        marginTop: '12px',
        lineHeight: 1.6
      }}>
        Encryption: <span style={{
          fontFamily: 'var(--font-mono)',
          color: 'var(--text-main)'
        }}>AES-256-GCM</span><br />
        Client-side verification active.
      </div>
    </section>
  );
};

const ActivityFeed = () => {
  const activities = [
    { type: 'UPLOAD', time: '10:42 AM', text: 'project-alpha-v2.zip', color: 'var(--text-tertiary)' },
    { type: 'SELF-HEAL', time: '09:15 AM', text: 'Repaired shard on OneDrive', color: 'var(--accent-secondary)' },
    { type: 'DOWNLOAD', time: '08:30 AM', text: 'database-backup.sql', color: 'var(--text-tertiary)' },
    { type: 'SYNC', time: '08:00 AM', text: 'Provider latency check', color: 'var(--text-tertiary)' }
  ];

  return (
    <section style={{
      gridColumn: '1 / 2',
      gridRow: '2 / 3',
      background: 'var(--bg-canvas)',
      border: '1px solid var(--border-color)',
      padding: '20px',
      position: 'relative',
      display: 'flex',
      flexDirection: 'column'
    }}>
      <CardCorners />
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'baseline',
        marginBottom: '20px'
      }}>
        <span style={{
          fontFamily: 'var(--font-mono)',
          fontSize: '11px',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
          color: 'var(--text-secondary)'
        }}>Recent Activity</span>
      </div>
      
      <ul style={{ listStyle: 'none', marginTop: '16px' }}>
        {activities.map((activity, index) => (
          <li
            key={index}
            style={{
              padding: '12px 0',
              borderBottom: index < activities.length - 1 ? '1px solid var(--grid-line)' : 'none',
              fontSize: '13px'
            }}
          >
            <div style={{
              display: 'flex',
              justifyContent: 'space-between',
              color: 'var(--text-tertiary)',
              fontSize: '11px',
              marginBottom: '4px',
              fontFamily: 'var(--font-mono)'
            }}>
              <span style={{ color: activity.color }}>{activity.type}</span>
              <span>{activity.time}</span>
            </div>
            <div>{activity.text}</div>
          </li>
        ))}
      </ul>
    </section>
  );
};

const TechSpecs = () => {
  const specs = [
    { label: 'Protocol', value: 'ZK-Rollup', color: 'var(--text-main)' },
    { label: 'Redundancy', value: '1.5x', color: 'var(--text-main)' },
    { label: 'Block Size', value: '4MB', color: 'var(--text-main)' },
    { label: 'Active Nodes', value: '142', color: 'var(--accent-primary)' },
    { label: 'API Version', value: 'v2.4.0', color: 'var(--text-main)' }
  ];

  return (
    <section style={{
      gridColumn: '3 / 4',
      gridRow: '2 / 3',
      background: 'var(--bg-canvas)',
      border: '1px solid var(--border-color)',
      padding: '20px',
      position: 'relative',
      display: 'flex',
      flexDirection: 'column'
    }}>
      <CardCorners />
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'baseline',
        marginBottom: '20px'
      }}>
        <span style={{
          fontFamily: 'var(--font-mono)',
          fontSize: '11px',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
          color: 'var(--text-secondary)'
        }}>Configuration</span>
      </div>
      
      {specs.map((spec, index) => (
        <div
          key={index}
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            padding: '12px 0',
            borderBottom: index < specs.length - 1 ? '1px solid var(--grid-line)' : 'none',
            fontSize: '13px'
          }}
        >
          <span style={{ color: 'var(--text-secondary)' }}>{spec.label}</span>
          <span style={{
            fontFamily: 'var(--font-mono)',
            color: spec.color
          }}>{spec.value}</span>
        </div>
      ))}
    </section>
  );
};

const Dashboard = () => (
  <>
    <StorageOverview />
    <ProviderMatrix />
    <SystemHealth />
    <ActivityFeed />
    <TechSpecs />
  </>
);

const App = () => {
  const [activeTab, setActiveTab] = useState('Dashboard');

  useEffect(() => {
    const styleContent = `
      * {
        box-sizing: border-box;
        margin: 0;
        padding: 0;
      }
      
      body {
        background-color: var(--bg-canvas);
        color: var(--text-main);
        font-family: var(--font-stack);
        -webkit-font-smoothing: antialiased;
        line-height: 1.5;
        height: 100vh;
        display: flex;
        flex-direction: column;
        overflow: hidden;
      }
      
      h1, h2, h3 {
        font-weight: 600;
        letter-spacing: -0.02em;
      }
      
      @media (max-width: 1200px) {
        main {
          grid-template-columns: 260px 1fr !important;
        }
      }
      
      @media (max-width: 900px) {
        main {
          grid-template-columns: 1fr !important;
        }
      }
    `;
    
    const style = document.createElement('style');
    style.textContent = styleContent;
    document.head.appendChild(style);
    
    return () => {
      document.head.removeChild(style);
    };
  }, []);

  return (
    <Router basename="/">
      <div style={customStyles.root}>
        <GridBackground />
        <Header activeTab={activeTab} setActiveTab={setActiveTab} />
        <main style={{
          flex: 1,
          padding: '24px',
          display: 'grid',
          gridTemplateColumns: '280px 1fr 320px',
          gridTemplateRows: 'auto 1fr',
          gap: '24px',
          overflowY: 'auto',
          maxWidth: '1600px',
          margin: '0 auto',
          width: '100%'
        }}>
          <Routes>
            <Route path="/" element={<Dashboard />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
};

export default App;