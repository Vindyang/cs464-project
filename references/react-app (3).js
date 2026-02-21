import React, { useState } from 'react';

const customStyles = {
  root: {
    '--bg-canvas': '#FFFFFF',
    '--bg-subtle': '#FAFAFA',
    '--text-main': '#111111',
    '--text-secondary': '#666666',
    '--text-tertiary': '#999999',
    '--accent-primary': '#004EEB',
    '--accent-primary-hover': '#003CC5',
    '--border-color': '#E5E5E5',
    '--grid-line': '#F5F5F5',
    '--font-stack': "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif",
    '--font-mono': "'SF Mono', 'Roboto Mono', 'Menlo', monospace",
    '--radius-sm': '2px',
    '--radius-md': '4px'
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
  brandIcon: {
    width: '24px',
    height: '24px',
    border: '1.5px solid var(--text-main)',
    position: 'relative'
  },
  brandIconAfter: {
    content: '',
    position: 'absolute',
    top: '50%',
    left: '50%',
    transform: 'translate(-50%, -50%)',
    width: '8px',
    height: '8px',
    background: 'var(--accent-primary)'
  },
  cardBefore: {
    content: '',
    position: 'absolute',
    width: '6px',
    height: '6px',
    borderColor: 'var(--text-tertiary)',
    borderStyle: 'solid',
    pointerEvents: 'none',
    opacity: 0.5,
    top: '-1px',
    left: '-1px',
    borderWidth: '1px 0 0 1px'
  },
  cardAfter: {
    content: '',
    position: 'absolute',
    width: '6px',
    height: '6px',
    borderColor: 'var(--text-tertiary)',
    borderStyle: 'solid',
    pointerEvents: 'none',
    opacity: 0.5,
    top: '-1px',
    right: '-1px',
    borderWidth: '1px 1px 0 0'
  }
};

const Header = () => {
  return (
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
      <div className="brand" style={{
        display: 'flex',
        alignItems: 'center',
        gap: '12px',
        fontWeight: 600,
        fontSize: '16px'
      }}>
        <div style={customStyles.brandIcon}>
          <div style={customStyles.brandIconAfter}></div>
        </div>
        <span>ZERO-STORE</span>
      </div>
      <nav style={{ display: 'flex', alignItems: 'center' }}>
        <a href="#" style={{
          color: 'var(--text-secondary)',
          textDecoration: 'none',
          fontSize: '14px',
          marginLeft: '24px'
        }}>Dashboard</a>
        <a href="#" style={{
          color: 'var(--text-secondary)',
          textDecoration: 'none',
          fontSize: '14px',
          marginLeft: '24px'
        }}>Nodes</a>
        <a href="#" style={{
          color: 'var(--text-secondary)',
          textDecoration: 'none',
          fontSize: '14px',
          marginLeft: '24px'
        }}>Settings</a>
        <a href="#" style={{
          color: 'var(--text-secondary)',
          textDecoration: 'none',
          fontSize: '14px',
          marginLeft: '24px'
        }}>API</a>
      </nav>
      <div>
        <a href="#" className="btn btn-outline" style={{
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
          background: 'transparent',
          border: '1px solid var(--border-color)',
          color: 'var(--text-main)'
        }}>Documentation</a>
      </div>
    </header>
  );
};

const StepItem = ({ number, title, description }) => {
  return (
    <div style={{
      display: 'flex',
      gap: '16px',
      marginBottom: '24px'
    }}>
      <div style={{
        width: '24px',
        height: '24px',
        border: '1px solid var(--border-color)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontFamily: 'var(--font-mono)',
        fontSize: '12px',
        flexShrink: 0,
        background: 'var(--bg-subtle)'
      }}>{number}</div>
      <div>
        <h4 style={{
          fontSize: '14px',
          fontWeight: 600,
          marginBottom: '4px'
        }}>{title}</h4>
        <p style={{
          fontSize: '13px',
          color: 'var(--text-secondary)'
        }}>{description}</p>
      </div>
    </div>
  );
};

const ProviderButton = ({ logo, name, onClick }) => {
  const [isHovered, setIsHovered] = useState(false);

  return (
    <a 
      href="#" 
      onClick={(e) => {
        e.preventDefault();
        if (onClick) onClick();
      }}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      style={{
        border: isHovered ? '1px solid var(--accent-primary)' : '1px solid var(--border-color)',
        padding: '16px',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: '8px',
        textDecoration: 'none',
        color: 'var(--text-main)',
        fontSize: '12px',
        fontWeight: 500,
        transition: 'all 0.2s',
        background: isHovered ? 'white' : 'var(--bg-subtle)'
      }}
    >
      <div style={{
        width: '32px',
        height: '32px',
        background: 'white',
        border: '1px solid var(--border-color)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontWeight: 700,
        fontSize: '10px'
      }}>{logo}</div>
      {name}
    </a>
  );
};

const Card = ({ children, style }) => {
  return (
    <section style={{
      background: 'var(--bg-canvas)',
      border: '1px solid var(--border-color)',
      padding: '32px',
      position: 'relative',
      ...style
    }}>
      <div style={customStyles.cardBefore}></div>
      <div style={customStyles.cardAfter}></div>
      {children}
    </section>
  );
};

const App = () => {
  const [selectedProvider, setSelectedProvider] = useState(null);

  const handleProviderClick = (provider) => {
    setSelectedProvider(provider);
  };

  const handleConnectProvider = () => {
    if (selectedProvider) {
      alert(`Connecting to ${selectedProvider}...`);
    } else {
      alert('Please select a provider first');
    }
  };

  return (
    <div style={{
      ...customStyles.root,
      minHeight: '100vh',
      display: 'flex',
      flexDirection: 'column',
      overflow: 'hidden',
      backgroundColor: 'var(--bg-canvas)',
      color: 'var(--text-main)',
      fontFamily: 'var(--font-stack)',
      WebkitFontSmoothing: 'antialiased'
    }}>
      <div style={customStyles.gridBg}></div>

      <Header />

      <main style={{
        flex: 1,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '40px'
      }}>
        <div style={{
          width: '100%',
          maxWidth: '840px',
          display: 'grid',
          gridTemplateColumns: '1.2fr 1fr',
          gap: '24px'
        }}>
          <Card>
            <span style={{
              fontFamily: 'var(--font-mono)',
              fontSize: '11px',
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
              color: 'var(--text-secondary)',
              marginBottom: '16px',
              display: 'block'
            }}>Getting Started</span>
            <h1 style={{
              fontSize: '32px',
              fontWeight: 600,
              letterSpacing: '-0.04em',
              marginBottom: '16px'
            }}>Initialize your distributed vault.</h1>
            <p style={{
              color: 'var(--text-secondary)',
              fontSize: '15px',
              lineHeight: 1.6,
              marginBottom: '32px'
            }}>
              Zero-Store splits your data into encrypted shards and distributes them across multiple cloud providers. To begin, you must connect at least two storage nodes.
            </p>

            <div>
              <StepItem 
                number="01"
                title="Connect Providers"
                description="Link AWS, GCP, or any S3-compatible storage to create your distribution network."
              />
              <StepItem 
                number="02"
                title="Configure Redundancy"
                description="Select your Reed-Solomon parity ratio to balance security and cost."
              />
              <StepItem 
                number="03"
                title="Upload First Object"
                description="Start storing files with client-side encryption and zero-knowledge privacy."
              />
            </div>

            <div style={{
              marginTop: '32px',
              paddingTop: '24px',
              borderTop: '1px solid var(--grid-line)',
              fontSize: '12px',
              color: 'var(--text-tertiary)'
            }}>
              System status: <span style={{ color: 'var(--accent-primary)' }}>WAITING_FOR_NODES</span>
            </div>
          </Card>

          <Card style={{ background: 'var(--bg-subtle)' }}>
            <span style={{
              fontFamily: 'var(--font-mono)',
              fontSize: '11px',
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
              color: 'var(--text-secondary)',
              marginBottom: '16px',
              display: 'block'
            }}>Step 1: Add Node</span>
            <h2 style={{ fontSize: '18px', marginBottom: '20px' }}>Select a Provider</h2>
            
            <div style={{
              display: 'grid',
              gridTemplateColumns: '1fr 1fr',
              gap: '12px',
              marginTop: '24px'
            }}>
              <ProviderButton logo="AWS" name="Amazon S3" onClick={() => handleProviderClick('Amazon S3')} />
              <ProviderButton logo="GCP" name="Google Drive" onClick={() => handleProviderClick('Google Drive')} />
              <ProviderButton logo="DBX" name="Dropbox" onClick={() => handleProviderClick('Dropbox')} />
              <ProviderButton logo="B2" name="Backblaze B2" onClick={() => handleProviderClick('Backblaze B2')} />
              <ProviderButton logo="MS" name="Azure Blob" onClick={() => handleProviderClick('Azure Blob')} />
              <ProviderButton logo="S3" name="Custom S3" onClick={() => handleProviderClick('Custom S3')} />
            </div>

            <div style={{ marginTop: '32px' }}>
              <a 
                href="#" 
                onClick={(e) => {
                  e.preventDefault();
                  handleConnectProvider();
                }}
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
                  border: 'none',
                  width: '100%'
                }}
              >Connect New Provider</a>
              <p style={{
                fontSize: '12px',
                color: 'var(--text-tertiary)',
                textAlign: 'center',
                marginTop: '16px'
              }}>
                Don't have a provider? <a href="#" style={{
                  color: 'var(--accent-primary)',
                  textDecoration: 'none'
                }}>Read our setup guide</a>
              </p>
            </div>
          </Card>
        </div>
      </main>
    </div>
  );
};

export default App;