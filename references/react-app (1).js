import React, { useState, useEffect } from 'react';

const App = () => {
  const [isModalOpen, setIsModalOpen] = useState(true);
  const [selectedNodes, setSelectedNodes] = useState(['AWS-US', 'GCP-GL', 'DBX-EU', 'ONE-US']);
  const [encryptionStandard, setEncryptionStandard] = useState('AES-256-GCM (Standard)');
  const [shardRedundancy, setShardRedundancy] = useState('9/6 (Standard Parity)');
  const [accessKey, setAccessKey] = useState('');

  useEffect(() => {
    const styleElement = document.createElement('style');
    styleElement.textContent = `
      :root {
        --bg-canvas: #FFFFFF;
        --bg-subtle: #FAFAFA;
        --text-main: #111111;
        --text-secondary: #666666;
        --text-tertiary: #999999;
        --accent-primary: #004EEB;
        --accent-primary-hover: #003CC5;
        --accent-secondary: #FF8866;
        --border-color: #E5E5E5;
        --grid-line: #F5F5F5;
        --font-stack: 'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
        --font-mono: 'SF Mono', 'Roboto Mono', 'Menlo', monospace;
        --radius-sm: 2px;
        --radius-md: 4px;
      }

      * { box-sizing: border-box; margin: 0; padding: 0; }

      body {
        background-color: var(--bg-canvas);
        color: var(--text-main);
        font-family: var(--font-stack);
        -webkit-font-smoothing: antialiased;
        height: 100vh;
        overflow: hidden;
      }

      @keyframes modalAppear {
        from { opacity: 0; transform: translateY(10px); }
        to { opacity: 1; transform: translateY(0); }
      }
    `;
    document.head.appendChild(styleElement);
    return () => document.head.removeChild(styleElement);
  }, []);

  const nodes = [
    'AWS-US', 'GCP-GL', 'DBX-EU', 'B2-WEST',
    'ONE-US', 'LOC-HUB', 'WAS-S3', 'AZR-VA'
  ];

  const toggleNode = (node) => {
    setSelectedNodes(prev =>
      prev.includes(node)
        ? prev.filter(n => n !== node)
        : [...prev, node]
    );
  };

  const handleInitializeSharding = () => {
    console.log('Initializing sharding with:', {
      encryptionStandard,
      shardRedundancy,
      selectedNodes,
      accessKey
    });
    setIsModalOpen(false);
  };

  return (
    <div style={{ height: '100vh', display: 'flex', flexDirection: 'column' }}>
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
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', fontWeight: 600 }}>
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
          <span>ZERO-STORE</span>
        </div>
        <nav>
          <a href="#" style={{ color: 'var(--text-secondary)', textDecoration: 'none', fontSize: '14px', marginLeft: '24px' }}>Dashboard</a>
          <a href="#" style={{ color: 'var(--text-secondary)', textDecoration: 'none', fontSize: '14px', marginLeft: '24px' }}>Nodes</a>
          <a href="#" style={{ color: 'var(--text-secondary)', textDecoration: 'none', fontSize: '14px', marginLeft: '24px' }}>Settings</a>
          <a href="#" style={{ color: 'var(--text-secondary)', textDecoration: 'none', fontSize: '14px', marginLeft: '24px' }}>API</a>
        </nav>
        <button
          onClick={() => setIsModalOpen(true)}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '0 20px',
            height: '40px',
            fontSize: '14px',
            fontWeight: 500,
            cursor: 'pointer',
            borderRadius: 'var(--radius-sm)',
            background: 'var(--accent-primary)',
            color: 'white',
            border: 'none'
          }}
        >
          Upload File
        </button>
      </header>

      <main style={{
        flex: 1,
        padding: '24px',
        display: 'grid',
        gridTemplateColumns: '280px 1fr 320px',
        gap: '24px',
        filter: isModalOpen ? 'blur(4px)' : 'none',
        pointerEvents: isModalOpen ? 'none' : 'auto'
      }}>
        <section style={{
          background: 'var(--bg-canvas)',
          border: '1px solid var(--border-color)',
          padding: '20px',
          position: 'relative'
        }}>
          <div style={{
            fontFamily: 'var(--font-mono)',
            fontSize: '11px',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            color: 'var(--text-secondary)'
          }}>Storage Used</div>
          <div style={{ fontSize: '32px', fontWeight: 600 }}>4.2TB</div>
        </section>
        <div style={{
          background: 'var(--bg-canvas)',
          border: '1px solid var(--border-color)',
          padding: '20px'
        }}>
          <div style={{
            fontFamily: 'var(--font-mono)',
            fontSize: '11px',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            color: 'var(--text-secondary)'
          }}>Providers</div>
          <div style={{ height: '200px', border: '1px solid var(--border-color)', marginTop: '10px' }}></div>
        </div>
        <section style={{
          background: 'var(--bg-canvas)',
          border: '1px solid var(--border-color)',
          padding: '20px',
          position: 'relative'
        }}>
          <div style={{
            fontFamily: 'var(--font-mono)',
            fontSize: '11px',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            color: 'var(--text-secondary)'
          }}>System Health</div>
        </section>
      </main>

      {isModalOpen && (
        <div style={{
          position: 'fixed',
          inset: 0,
          background: 'rgba(255, 255, 255, 0.85)',
          backdropFilter: 'blur(2px)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 100
        }}>
          <div style={{
            width: '640px',
            background: 'var(--bg-canvas)',
            border: '1px solid var(--text-main)',
            boxShadow: '0 20px 40px rgba(0,0,0,0.1)',
            position: 'relative',
            animation: 'modalAppear 0.3s cubic-bezier(0.16, 1, 0.3, 1)'
          }}>
            <div style={{ position: 'absolute', top: '-1px', left: '-1px', width: '8px', height: '8px', borderTop: '1px solid var(--text-main)', borderLeft: '1px solid var(--text-main)' }} />
            <div style={{ position: 'absolute', top: '-1px', right: '-1px', width: '8px', height: '8px', borderTop: '1px solid var(--text-main)', borderRight: '1px solid var(--text-main)' }} />
            <div style={{ position: 'absolute', bottom: '-1px', left: '-1px', width: '8px', height: '8px', borderBottom: '1px solid var(--text-main)', borderLeft: '1px solid var(--text-main)' }} />
            <div style={{ position: 'absolute', bottom: '-1px', right: '-1px', width: '8px', height: '8px', borderBottom: '1px solid var(--text-main)', borderRight: '1px solid var(--text-main)' }} />

            <div style={{
              padding: '24px',
              borderBottom: '1px solid var(--border-color)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center'
            }}>
              <h2 style={{ fontSize: '18px', fontWeight: 600 }}>New File Upload</h2>
              <div style={{
                fontFamily: 'var(--font-mono)',
                fontSize: '11px',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                color: 'var(--accent-primary)'
              }}>Secure Session</div>
            </div>

            <div style={{ padding: '24px' }}>
              <div style={{
                border: '1px dashed var(--border-color)',
                background: 'var(--bg-subtle)',
                padding: '40px',
                textAlign: 'center',
                marginBottom: '24px',
                cursor: 'pointer'
              }}>
                <div style={{
                  fontFamily: 'var(--font-mono)',
                  fontSize: '11px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  marginBottom: '8px',
                  color: 'var(--accent-primary)'
                }}>Drag & Drop</div>
                <div style={{ fontSize: '14px', color: 'var(--text-main)' }}>Click to browse or drop files here</div>
                <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginTop: '8px' }}>Maximum single file size: 2.0 GB</div>
              </div>

              <div style={{
                display: 'grid',
                gridTemplateColumns: '1fr 1fr',
                gap: '24px',
                marginBottom: '24px'
              }}>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                  <label style={{
                    fontFamily: 'var(--font-mono)',
                    fontSize: '11px',
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                    color: 'var(--text-secondary)'
                  }}>Encryption Standard</label>
                  <select
                    value={encryptionStandard}
                    onChange={(e) => setEncryptionStandard(e.target.value)}
                    style={{
                      height: '36px',
                      border: '1px solid var(--border-color)',
                      padding: '0 12px',
                      fontFamily: 'var(--font-stack)',
                      fontSize: '13px',
                      borderRadius: 'var(--radius-sm)',
                      background: 'white',
                      outline: 'none'
                    }}
                  >
                    <option>AES-256-GCM (Standard)</option>
                    <option>ChaCha20-Poly1305</option>
                    <option>RSA-4096 (Vault only)</option>
                  </select>
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                  <label style={{
                    fontFamily: 'var(--font-mono)',
                    fontSize: '11px',
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                    color: 'var(--text-secondary)'
                  }}>Shard Redundancy</label>
                  <select
                    value={shardRedundancy}
                    onChange={(e) => setShardRedundancy(e.target.value)}
                    style={{
                      height: '36px',
                      border: '1px solid var(--border-color)',
                      padding: '0 12px',
                      fontFamily: 'var(--font-stack)',
                      fontSize: '13px',
                      borderRadius: 'var(--radius-sm)',
                      background: 'white',
                      outline: 'none'
                    }}
                  >
                    <option>9/6 (Standard Parity)</option>
                    <option>12/8 (High Availability)</option>
                    <option>4/2 (Low Latency)</option>
                  </select>
                </div>
              </div>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', marginBottom: '24px' }}>
                <label style={{
                  fontFamily: 'var(--font-mono)',
                  fontSize: '11px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  color: 'var(--text-secondary)'
                }}>Preferred Distribution</label>
                <div style={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(4, 1fr)',
                  gap: '8px',
                  marginTop: '4px'
                }}>
                  {nodes.map(node => (
                    <div
                      key={node}
                      onClick={() => toggleNode(node)}
                      style={{
                        border: selectedNodes.includes(node) ? '1px solid var(--accent-primary)' : '1px solid var(--border-color)',
                        background: selectedNodes.includes(node) ? 'rgba(0, 78, 235, 0.05)' : 'transparent',
                        color: selectedNodes.includes(node) ? 'var(--accent-primary)' : 'var(--text-main)',
                        padding: '8px',
                        fontSize: '10px',
                        textAlign: 'center',
                        fontFamily: 'var(--font-mono)',
                        cursor: 'pointer'
                      }}
                    >
                      {node}
                    </div>
                  ))}
                </div>
                <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginTop: '8px' }}>
                  Minimum 4 providers required for selected redundancy.
                </div>
              </div>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                <label style={{
                  fontFamily: 'var(--font-mono)',
                  fontSize: '11px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  color: 'var(--text-secondary)'
                }}>Private Access Key (Optional)</label>
                <input
                  type="password"
                  placeholder="Leave blank for auto-generated key"
                  value={accessKey}
                  onChange={(e) => setAccessKey(e.target.value)}
                  style={{
                    height: '36px',
                    border: '1px solid var(--border-color)',
                    padding: '0 12px',
                    fontFamily: 'var(--font-stack)',
                    fontSize: '13px',
                    borderRadius: 'var(--radius-sm)',
                    background: 'white',
                    outline: 'none'
                  }}
                />
              </div>
            </div>

            <div style={{
              padding: '20px 24px',
              borderTop: '1px solid var(--border-color)',
              display: 'flex',
              justifyContent: 'flex-end',
              gap: '12px',
              background: 'var(--bg-subtle)'
            }}>
              <button
                onClick={() => setIsModalOpen(false)}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  padding: '0 20px',
                  height: '40px',
                  fontSize: '14px',
                  fontWeight: 500,
                  cursor: 'pointer',
                  borderRadius: 'var(--radius-sm)',
                  background: 'transparent',
                  border: '1px solid transparent',
                  color: 'var(--text-secondary)'
                }}
              >
                Cancel
              </button>
              <button
                onClick={handleInitializeSharding}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  padding: '0 32px',
                  height: '40px',
                  fontSize: '14px',
                  fontWeight: 500,
                  cursor: 'pointer',
                  borderRadius: 'var(--radius-sm)',
                  background: 'var(--accent-primary)',
                  color: 'white',
                  border: 'none'
                }}
              >
                Initialize Sharding
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default App;