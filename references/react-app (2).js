import React, { useState } from 'react';

const App = () => {
  const [isModalOpen, setIsModalOpen] = useState(true);
  const [selectedNodes, setSelectedNodes] = useState(['AWS-US', 'GCP-GL', 'DBX-EU', 'ONE-US']);
  const [encryptionStandard, setEncryptionStandard] = useState('AES-256-GCM (Standard)');
  const [shardRedundancy, setShardRedundancy] = useState('9/6 (Standard Parity)');
  const [privateKey, setPrivateKey] = useState('');
  const [isDragging, setIsDragging] = useState(false);

  const nodes = ['AWS-US', 'GCP-GL', 'DBX-EU', 'B2-WEST', 'ONE-US', 'LOC-HUB', 'WAS-S3', 'AZR-VA'];

  const toggleNode = (node) => {
    if (selectedNodes.includes(node)) {
      setSelectedNodes(selectedNodes.filter(n => n !== node));
    } else {
      setSelectedNodes([...selectedNodes, node]);
    }
  };

  const handleDragOver = (e) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = () => {
    setIsDragging(false);
  };

  const handleDrop = (e) => {
    e.preventDefault();
    setIsDragging(false);
  };

  const handleFileSelect = () => {
    console.log('File browser opened');
  };

  const handleInitialize = () => {
    console.log('Initializing sharding with:', {
      encryptionStandard,
      shardRedundancy,
      selectedNodes,
      privateKey
    });
  };

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
      '--radius-md': '4px'
    },
    body: {
      backgroundColor: 'var(--bg-canvas)',
      color: 'var(--text-main)',
      fontFamily: 'var(--font-stack)',
      WebkitFontSmoothing: 'antialiased',
      height: '100vh',
      display: 'flex',
      flexDirection: 'column',
      overflow: 'hidden',
      margin: 0,
      padding: 0
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
      fontWeight: 600
    },
    brandIcon: {
      width: '24px',
      height: '24px',
      border: '1.5px solid var(--text-main)',
      position: 'relative'
    },
    brandIconAfter: {
      position: 'absolute',
      top: '50%',
      left: '50%',
      transform: 'translate(-50%, -50%)',
      width: '8px',
      height: '8px',
      background: 'var(--accent-primary)'
    },
    nav: {
      display: 'flex',
      alignItems: 'center'
    },
    navLink: {
      color: 'var(--text-secondary)',
      textDecoration: 'none',
      fontSize: '14px',
      marginLeft: '24px'
    },
    main: {
      flex: 1,
      padding: '24px',
      display: 'grid',
      gridTemplateColumns: '280px 1fr 320px',
      gap: '24px',
      filter: isModalOpen ? 'blur(4px)' : 'none',
      pointerEvents: isModalOpen ? 'none' : 'auto'
    },
    card: {
      background: 'var(--bg-canvas)',
      border: '1px solid var(--border-color)',
      padding: '20px',
      position: 'relative'
    },
    labelMono: {
      fontFamily: 'var(--font-mono)',
      fontSize: '11px',
      textTransform: 'uppercase',
      letterSpacing: '0.05em',
      color: 'var(--text-secondary)'
    },
    modalOverlay: {
      position: 'fixed',
      inset: 0,
      background: 'rgba(255, 255, 255, 0.85)',
      backdropFilter: 'blur(2px)',
      display: isModalOpen ? 'flex' : 'none',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 100
    },
    modal: {
      width: '640px',
      background: 'var(--bg-canvas)',
      border: '1px solid var(--text-main)',
      boxShadow: '0 20px 40px rgba(0,0,0,0.1)',
      position: 'relative',
      animation: 'modalAppear 0.3s cubic-bezier(0.16, 1, 0.3, 1)'
    },
    modalHeader: {
      padding: '24px',
      borderBottom: '1px solid var(--border-color)',
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center'
    },
    modalBody: {
      padding: '24px'
    },
    dropZone: {
      border: isDragging ? '1px dashed var(--accent-primary)' : '1px dashed var(--border-color)',
      background: isDragging ? 'rgba(0, 78, 235, 0.02)' : 'var(--bg-subtle)',
      padding: '40px',
      textAlign: 'center',
      marginBottom: '24px',
      transition: 'all 0.2s',
      cursor: 'pointer'
    },
    configGrid: {
      display: 'grid',
      gridTemplateColumns: '1fr 1fr',
      gap: '24px',
      marginBottom: '24px'
    },
    configItem: {
      display: 'flex',
      flexDirection: 'column',
      gap: '8px'
    },
    select: {
      height: '36px',
      border: '1px solid var(--border-color)',
      padding: '0 12px',
      fontFamily: 'var(--font-stack)',
      fontSize: '13px',
      borderRadius: 'var(--radius-sm)',
      background: 'white',
      outline: 'none'
    },
    input: {
      height: '36px',
      border: '1px solid var(--border-color)',
      padding: '0 12px',
      fontFamily: 'var(--font-stack)',
      fontSize: '13px',
      borderRadius: 'var(--radius-sm)',
      background: 'white',
      outline: 'none'
    },
    distributionSelector: {
      display: 'grid',
      gridTemplateColumns: 'repeat(4, 1fr)',
      gap: '8px',
      marginTop: '4px'
    },
    distNode: {
      border: '1px solid var(--border-color)',
      padding: '8px',
      fontSize: '10px',
      textAlign: 'center',
      fontFamily: 'var(--font-mono)',
      cursor: 'pointer'
    },
    distNodeActive: {
      borderColor: 'var(--accent-primary)',
      background: 'rgba(0, 78, 235, 0.05)',
      color: 'var(--accent-primary)'
    },
    modalFooter: {
      padding: '20px 24px',
      borderTop: '1px solid var(--border-color)',
      display: 'flex',
      justifyContent: 'flex-end',
      gap: '12px',
      background: 'var(--bg-subtle)'
    },
    btn: {
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '0 20px',
      height: '40px',
      fontSize: '14px',
      fontWeight: 500,
      cursor: 'pointer',
      borderRadius: 'var(--radius-sm)',
      textDecoration: 'none',
      transition: 'all 0.2s'
    },
    btnPrimary: {
      background: 'var(--accent-primary)',
      color: 'white',
      border: 'none'
    },
    btnGhost: {
      background: 'transparent',
      border: '1px solid transparent',
      color: 'var(--text-secondary)'
    },
    limitTag: {
      fontSize: '11px',
      color: 'var(--text-tertiary)',
      marginTop: '8px'
    },
    cornerMark: {
      position: 'absolute',
      width: '8px',
      height: '8px',
      borderColor: 'var(--text-main)',
      borderStyle: 'solid',
      pointerEvents: 'none'
    }
  };

  return (
    <div style={customStyles.root}>
      <style>{`
        @keyframes modalAppear {
          from { opacity: 0; transform: translateY(10px); }
          to { opacity: 1; transform: translateY(0); }
        }
      `}</style>
      <div style={customStyles.body}>
        <div style={customStyles.gridBg}></div>

        <header style={customStyles.header}>
          <div style={customStyles.brand}>
            <div style={customStyles.brandIcon}>
              <div style={customStyles.brandIconAfter}></div>
            </div>
            <span>ZERO-STORE</span>
          </div>
          <nav style={customStyles.nav}>
            <a href="#" style={{...customStyles.navLink, color: 'var(--text-main)'}}>Dashboard</a>
            <a href="#" style={customStyles.navLink}>Nodes</a>
            <a href="#" style={customStyles.navLink}>Settings</a>
            <a href="#" style={customStyles.navLink}>API</a>
          </nav>
          <button 
            style={{...customStyles.btn, ...customStyles.btnPrimary}}
            onClick={() => setIsModalOpen(true)}
          >
            Upload File
          </button>
        </header>

        <main style={customStyles.main}>
          <section style={customStyles.card}>
            <div style={customStyles.labelMono}>Storage Used</div>
            <div style={{fontSize: '32px', fontWeight: 600}}>4.2TB</div>
          </section>
          <div style={customStyles.card}>
            <div style={customStyles.labelMono}>Providers</div>
            <div style={{height: '200px', border: '1px solid var(--border-color)', marginTop: '10px'}}></div>
          </div>
          <section style={customStyles.card}>
            <div style={customStyles.labelMono}>System Health</div>
          </section>
        </main>

        <div style={customStyles.modalOverlay}>
          <div style={customStyles.modal}>
            <div style={{...customStyles.cornerMark, top: '-1px', left: '-1px', borderWidth: '1px 0 0 1px'}}></div>
            <div style={{...customStyles.cornerMark, top: '-1px', right: '-1px', borderWidth: '1px 1px 0 0'}}></div>
            <div style={{...customStyles.cornerMark, bottom: '-1px', left: '-1px', borderWidth: '0 0 1px 1px'}}></div>
            <div style={{...customStyles.cornerMark, bottom: '-1px', right: '-1px', borderWidth: '0 1px 1px 0'}}></div>
            
            <div style={customStyles.modalHeader}>
              <h2 style={{fontSize: '18px', fontWeight: 600}}>New File Upload</h2>
              <div style={{...customStyles.labelMono, color: 'var(--accent-primary)'}}>Secure Session</div>
            </div>

            <div style={customStyles.modalBody}>
              <div 
                style={customStyles.dropZone}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
                onClick={handleFileSelect}
              >
                <div style={{...customStyles.labelMono, marginBottom: '8px', color: 'var(--accent-primary)'}}>Drag & Drop</div>
                <div style={{fontSize: '14px', color: 'var(--text-main)'}}>Click to browse or drop files here</div>
                <div style={customStyles.limitTag}>Maximum single file size: 2.0 GB</div>
              </div>

              <div style={customStyles.configGrid}>
                <div style={customStyles.configItem}>
                  <label style={customStyles.labelMono}>Encryption Standard</label>
                  <select 
                    style={customStyles.select}
                    value={encryptionStandard}
                    onChange={(e) => setEncryptionStandard(e.target.value)}
                  >
                    <option>AES-256-GCM (Standard)</option>
                    <option>ChaCha20-Poly1305</option>
                    <option>RSA-4096 (Vault only)</option>
                  </select>
                </div>
                <div style={customStyles.configItem}>
                  <label style={customStyles.labelMono}>Shard Redundancy</label>
                  <select 
                    style={customStyles.select}
                    value={shardRedundancy}
                    onChange={(e) => setShardRedundancy(e.target.value)}
                  >
                    <option>9/6 (Standard Parity)</option>
                    <option>12/8 (High Availability)</option>
                    <option>4/2 (Low Latency)</option>
                  </select>
                </div>
              </div>

              <div style={{...customStyles.configItem, marginBottom: '24px'}}>
                <label style={customStyles.labelMono}>Preferred Distribution</label>
                <div style={customStyles.distributionSelector}>
                  {nodes.map((node) => (
                    <div 
                      key={node}
                      style={{
                        ...customStyles.distNode,
                        ...(selectedNodes.includes(node) ? customStyles.distNodeActive : {})
                      }}
                      onClick={() => toggleNode(node)}
                    >
                      {node}
                    </div>
                  ))}
                </div>
                <div style={customStyles.limitTag}>Minimum 4 providers required for selected redundancy.</div>
              </div>

              <div style={customStyles.configItem}>
                <label style={customStyles.labelMono}>Private Access Key (Optional)</label>
                <input 
                  type="password" 
                  placeholder="Leave blank for auto-generated key"
                  style={customStyles.input}
                  value={privateKey}
                  onChange={(e) => setPrivateKey(e.target.value)}
                />
              </div>
            </div>

            <div style={customStyles.modalFooter}>
              <button 
                style={{...customStyles.btn, ...customStyles.btnGhost}}
                onClick={() => setIsModalOpen(false)}
              >
                Cancel
              </button>
              <button 
                style={{...customStyles.btn, ...customStyles.btnPrimary, padding: '0 32px'}}
                onClick={handleInitialize}
              >
                Initialize Sharding
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default App;