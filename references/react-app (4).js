import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Link, useNavigate } from 'react-router-dom';

const GridBackground = () => (
  <div style={{
    position: 'absolute',
    top: 0,
    left: 0,
    width: '100%',
    height: '100%',
    backgroundImage: 'linear-gradient(#F5F5F5 1px, transparent 1px), linear-gradient(90deg, #F5F5F5 1px, transparent 1px)',
    backgroundSize: '40px 40px',
    zIndex: -1,
    pointerEvents: 'none'
  }} />
);

const BrandIcon = () => (
  <div style={{
    width: '24px',
    height: '24px',
    border: '1.5px solid #111111',
    position: 'relative'
  }}>
    <div style={{
      position: 'absolute',
      top: '50%',
      left: '50%',
      transform: 'translate(-50%, -50%)',
      width: '8px',
      height: '8px',
      background: '#004EEB'
    }} />
  </div>
);

const Card = ({ children, className = '' }) => (
  <div style={{
    background: '#FFFFFF',
    border: '1px solid #E5E5E5',
    padding: '20px',
    position: 'relative',
    display: 'flex',
    flexDirection: 'column'
  }} className={className}>
    <div style={{
      position: 'absolute',
      top: '-1px',
      left: '-1px',
      width: '6px',
      height: '6px',
      borderTop: '1px solid #999999',
      borderLeft: '1px solid #999999',
      opacity: 0.5
    }} />
    <div style={{
      position: 'absolute',
      top: '-1px',
      right: '-1px',
      width: '6px',
      height: '6px',
      borderTop: '1px solid #999999',
      borderRight: '1px solid #999999',
      opacity: 0.5
    }} />
    <div style={{
      position: 'absolute',
      bottom: '-1px',
      left: '-1px',
      width: '6px',
      height: '6px',
      borderBottom: '1px solid #999999',
      borderLeft: '1px solid #999999',
      opacity: 0.5
    }} />
    <div style={{
      position: 'absolute',
      bottom: '-1px',
      right: '-1px',
      width: '6px',
      height: '6px',
      borderBottom: '1px solid #999999',
      borderRight: '1px solid #999999',
      opacity: 0.5
    }} />
    {children}
  </div>
);

const Header = () => {
  const navigate = useNavigate();
  const [activeNav, setActiveNav] = useState('nodes');

  return (
    <header style={{
      height: '64px',
      borderBottom: '1px solid #E5E5E5',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      padding: '0 24px',
      background: '#FFFFFF',
      zIndex: 10
    }}>
      <Link to="/" style={{
        display: 'flex',
        alignItems: 'center',
        gap: '12px',
        fontWeight: 600,
        fontSize: '16px',
        textDecoration: 'none',
        color: '#111111'
      }}>
        <BrandIcon />
        <span>ZERO-STORE</span>
      </Link>
      <nav>
        {[
          { id: 'dashboard', label: 'Dashboard', path: '/' },
          { id: 'nodes', label: 'Nodes', path: '/nodes' },
          { id: 'settings', label: 'Settings', path: '/settings' },
          { id: 'api', label: 'API', path: '/api' }
        ].map(item => (
          <Link
            key={item.id}
            to={item.path}
            onClick={() => setActiveNav(item.id)}
            style={{
              color: activeNav === item.id ? '#111111' : '#666666',
              textDecoration: 'none',
              fontSize: '14px',
              marginLeft: '24px',
              transition: 'color 0.2s'
            }}
          >
            {item.label}
          </Link>
        ))}
      </nav>
      <div style={{ display: 'flex', gap: '12px' }}>
        <button style={{
          display: 'inline-flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '0 16px',
          height: '36px',
          fontSize: '13px',
          fontWeight: 500,
          cursor: 'pointer',
          transition: 'all 0.2s',
          borderRadius: '2px',
          backgroundColor: '#004EEB',
          color: 'white',
          border: 'none'
        }}>
          Upload File
        </button>
      </div>
    </header>
  );
};

const ProviderDetails = () => {
  const [hoveredBar, setHoveredBar] = useState(null);

  const chartData = [40, 60, 35, 85, 95, 70, 45, 30, 55, 80, 65, 50, 90, 75, 40, 60, 35, 85, 25, 45, 65, 55, 70, 30];

  return (
    <main style={{
      flex: 1,
      padding: '24px',
      display: 'grid',
      gridTemplateColumns: '320px 1fr',
      gridTemplateRows: 'auto 1fr',
      gap: '24px',
      overflowY: 'auto',
      maxWidth: '1400px',
      margin: '0 auto',
      width: '100%'
    }}>
      <div style={{
        gridColumn: '1 / -1',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        marginBottom: '8px'
      }}>
        <div>
          <div style={{
            display: 'flex',
            gap: '8px',
            fontSize: '14px',
            color: '#666666',
            marginBottom: '8px'
          }}>
            <span>Providers</span>
            <span>/</span>
            <span style={{ color: '#111111' }}>Amazon S3 (us-east-1)</span>
          </div>
          <h1 style={{
            fontSize: '24px',
            display: 'flex',
            alignItems: 'center',
            gap: '12px'
          }}>
            <span style={{
              background: '#F0F0F0',
              padding: '4px 8px',
              fontSize: '14px',
              borderRadius: '2px'
            }}>AWS</span>
            Amazon S3
          </h1>
        </div>
        <div style={{ display: 'flex', gap: '12px' }}>
          <button style={{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '0 16px',
            height: '36px',
            fontSize: '13px',
            fontWeight: 500,
            cursor: 'pointer',
            transition: 'all 0.2s',
            borderRadius: '2px',
            background: 'transparent',
            border: '1px solid #E5E5E5',
            color: '#111111'
          }}>
            Edit Credentials
          </button>
          <button style={{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '0 16px',
            height: '36px',
            fontSize: '13px',
            fontWeight: 500,
            cursor: 'pointer',
            transition: 'all 0.2s',
            borderRadius: '2px',
            background: 'transparent',
            border: '1px solid #FFCDCD',
            color: '#DC2626'
          }}>
            Detach Provider
          </button>
        </div>
      </div>

      <aside style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
        <Card>
          <span style={{
            fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
            fontSize: '11px',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            color: '#666666'
          }}>Operational Status</span>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            marginTop: '12px'
          }}>
            <div style={{
              width: '8px',
              height: '8px',
              background: '#10B981',
              borderRadius: '50%'
            }} />
            <span style={{ fontWeight: 500 }}>Active & Syncing</span>
          </div>
          <div style={{ marginTop: '16px' }}>
            <span style={{
              fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
              fontSize: '11px',
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
              color: '#666666'
            }}>Current Load</span>
            <div style={{
              fontSize: '28px',
              fontWeight: 600,
              letterSpacing: '-0.02em'
            }}>
              2.14 <span style={{ fontSize: '14px', color: '#999999' }}>TB</span>
            </div>
            <div style={{
              fontSize: '12px',
              color: '#666666',
              marginTop: '4px'
            }}>Quota: 5.00 TB (43%)</div>
          </div>
        </Card>

        <Card>
          <span style={{
            fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
            fontSize: '11px',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            color: '#666666'
          }}>Cost Analysis (MTD)</span>
          <div style={{
            fontSize: '28px',
            fontWeight: 600,
            letterSpacing: '-0.02em',
            marginTop: '12px'
          }}>$42.80</div>
          <div style={{
            fontSize: '12px',
            color: '#666666',
            marginTop: '4px'
          }}>Est. Monthly: $58.12</div>
          
          <div style={{ marginTop: '20px' }}>
            <div style={{
              display: 'flex',
              justifyContent: 'space-between',
              padding: '12px 0',
              borderBottom: '1px solid #F5F5F5',
              fontSize: '13px'
            }}>
              <span style={{
                fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
                fontSize: '10px',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                color: '#666666'
              }}>Storage Cost</span>
              <span style={{ fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace" }}>$34.12</span>
            </div>
            <div style={{
              display: 'flex',
              justifyContent: 'space-between',
              padding: '12px 0',
              borderBottom: '1px solid #F5F5F5',
              fontSize: '13px'
            }}>
              <span style={{
                fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
                fontSize: '10px',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                color: '#666666'
              }}>Egress</span>
              <span style={{ fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace" }}>$8.68</span>
            </div>
          </div>
        </Card>

        <Card>
          <span style={{
            fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
            fontSize: '11px',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            color: '#666666'
          }}>Provider Settings</span>
          <div style={{ marginTop: '12px' }}>
            {[
              { label: 'Storage Class', value: 'Standard', color: '#111111' },
              { label: 'Versioning', value: 'Enabled', color: '#10B981' },
              { label: 'Server Encryption', value: 'SSE-S3', color: '#111111' },
              { label: 'Lifecycle Policy', value: 'Active', color: '#004EEB' }
            ].map((item, idx) => (
              <div key={idx} style={{
                display: 'flex',
                justifyContent: 'space-between',
                padding: '12px 0',
                borderBottom: idx < 3 ? '1px solid #F5F5F5' : 'none',
                fontSize: '13px'
              }}>
                <span style={{ color: '#666666', fontSize: '13px' }}>{item.label}</span>
                <span style={{
                  fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
                  color: item.color
                }}>{item.value}</span>
              </div>
            ))}
          </div>
        </Card>
      </aside>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
        <Card>
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center'
          }}>
            <span style={{
              fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
              fontSize: '11px',
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
              color: '#666666'
            }}>Bandwidth Usage (24h)</span>
            <div style={{ display: 'flex', gap: '16px', fontSize: '11px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                <div style={{ width: '8px', height: '8px', background: '#004EEB' }} />
                <span style={{
                  fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
                  fontSize: '11px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  color: '#666666'
                }}>Ingress</span>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                <div style={{ width: '8px', height: '8px', background: '#FF8866' }} />
                <span style={{
                  fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
                  fontSize: '11px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  color: '#666666'
                }}>Egress</span>
              </div>
            </div>
          </div>
          
          <div style={{
            height: '200px',
            width: '100%',
            marginTop: '20px',
            position: 'relative',
            display: 'flex',
            alignItems: 'flex-end',
            gap: '4px'
          }}>
            {chartData.map((height, idx) => (
              <div
                key={idx}
                onMouseEnter={() => setHoveredBar(idx)}
                onMouseLeave={() => setHoveredBar(null)}
                style={{
                  flex: 1,
                  background: '#004EEB',
                  opacity: hoveredBar === idx ? 1 : 0.8,
                  minHeight: '4px',
                  height: `${height}%`,
                  transition: 'opacity 0.2s',
                  cursor: 'pointer'
                }}
              />
            ))}
          </div>
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            marginTop: '12px'
          }}>
            {['00:00', '12:00', '23:59'].map(time => (
              <span key={time} style={{
                fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
                fontSize: '9px',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                color: '#666666'
              }}>{time}</span>
            ))}
          </div>
        </Card>

        <Card>
          <span style={{
            fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
            fontSize: '11px',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            color: '#666666',
            marginBottom: '24px'
          }}>File Type Distribution</span>
          
          {[
            { label: 'Media (.mp4, .jpg)', width: 65, size: '1.39 TB', color: '#004EEB' },
            { label: 'Archives (.zip, .gz)', width: 25, size: '535 GB', color: '#FF8866' },
            { label: 'Documents (.pdf)', width: 8, size: '171 GB', color: '#004EEB', opacity: 0.5 },
            { label: 'Other', width: 2, size: '42 GB', color: '#004EEB', opacity: 0.2 }
          ].map((item, idx) => (
            <div key={idx} style={{
              display: 'flex',
              alignItems: 'center',
              gap: '12px',
              marginBottom: '16px'
            }}>
              <div style={{
                width: '100px',
                fontSize: '13px',
                color: '#666666'
              }}>{item.label}</div>
              <div style={{
                flex: 1,
                height: '6px',
                background: '#F5F5F5'
              }}>
                <div style={{
                  height: '100%',
                  width: `${item.width}%`,
                  background: item.color,
                  opacity: item.opacity || 1
                }} />
              </div>
              <div style={{
                fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
                width: '60px',
                textAlign: 'right',
                fontSize: '12px'
              }}>{item.size}</div>
            </div>
          ))}
        </Card>

        <Card>
          <span style={{
            fontFamily: "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
            fontSize: '11px',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            color: '#666666',
            marginBottom: '16px'
          }}>Local Shard Integrity</span>
          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(20, 1fr)',
            gap: '4px'
          }}>
            {Array.from({ length: 20 }).map((_, idx) => (
              <div key={idx} style={{
                aspectRatio: '1',
                background: idx === 0 ? '#004EEB' : idx === 5 ? '#FF8866' : '#004EEB',
                opacity: idx === 0 ? 0.1 : 0.8
              }} />
            ))}
          </div>
          <div style={{
            marginTop: '12px',
            fontSize: '11px',
            color: '#999999'
          }}>
            Showing 128 of 4,092 active shards. All shards verified 6m ago.
          </div>
        </Card>
      </div>
    </main>
  );
};

const App = () => {
  return (
    <Router basename="/">
      <div style={{
        backgroundColor: '#FFFFFF',
        color: '#111111',
        fontFamily: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif",
        WebkitFontSmoothing: 'antialiased',
        lineHeight: 1.5,
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
        position: 'relative'
      }}>
        <GridBackground />
        <Header />
        <Routes>
          <Route path="/" element={<ProviderDetails />} />
          <Route path="/nodes" element={<ProviderDetails />} />
          <Route path="/settings" element={<ProviderDetails />} />
          <Route path="/api" element={<ProviderDetails />} />
        </Routes>
      </div>
    </Router>
  );
};

export default App;