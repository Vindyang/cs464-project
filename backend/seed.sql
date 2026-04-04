-- Seed data for Omnishard demo
-- User ID: 30o2NG7odPPj7M9a7kScKMDsKaiBpnm2
-- Run AFTER the schema/init-local.sql has been applied.
-- Safe to re-run: all inserts use ON CONFLICT DO NOTHING.

-- Ensure user_id column exists for running DBs created before this column was added.
ALTER TABLE files ADD COLUMN IF NOT EXISTS user_id TEXT;
CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);

-- ============================================================
-- FILES  (n=6, k=4 per file — 4 data shards + 2 parity shards)
-- ============================================================

INSERT INTO files (id, user_id, original_name, original_size, total_chunks, n, k, shard_size, status)
VALUES
  ('a1b2c3d4-0001-0000-0000-000000000001', '30o2NG7odPPj7M9a7kScKMDsKaiBpnm2', 'project_backup_v2.zip',       2684354560, 1, 6, 4, 447392426, 'UPLOADED'),
  ('a1b2c3d4-0002-0000-0000-000000000002', '30o2NG7odPPj7M9a7kScKMDsKaiBpnm2', 'database_export_2026.sql',    2040109465, 1, 6, 4, 340018244, 'UPLOADED'),
  ('a1b2c3d4-0003-0000-0000-000000000003', '30o2NG7odPPj7M9a7kScKMDsKaiBpnm2', 'sensitive_docs_encrypted.pdf',  15728640, 1, 6, 4,   2621440, 'DEGRADED'),
  ('a1b2c3d4-0004-0000-0000-000000000004', '30o2NG7odPPj7M9a7kScKMDsKaiBpnm2', 'family_photos_2023.tar.gz',  4831838208, 1, 6, 4, 805306368, 'DEGRADED'),
  ('a1b2c3d4-0005-0000-0000-000000000005', '30o2NG7odPPj7M9a7kScKMDsKaiBpnm2', 'quarterly_report_Q4.xlsx',     49283072, 1, 6, 4,   8213845, 'UPLOADED')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- SHARDS
-- UPLOADED files: 6 HEALTHY shards  (indices 0-5)
-- DEGRADED files: 4 HEALTHY + 2 MISSING shards (indices 4-5 missing)
-- Provider distribution: alternating googleDrive / awsS3
-- ============================================================

-- project_backup_v2.zip (UPLOADED — all 6 healthy)
INSERT INTO shards (id, file_id, chunk_index, shard_index, shard_type, remote_id, provider, checksum_sha256, status)
VALUES
  ('b0000001-0001-0000-0000-000000000001', 'a1b2c3d4-0001-0000-0000-000000000001', 0, 0, 'DATA',   'gdrive-seed-001-s0', 'googleDrive', 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa', 'HEALTHY'),
  ('b0000001-0001-0000-0000-000000000002', 'a1b2c3d4-0001-0000-0000-000000000001', 0, 1, 'DATA',   'awss3-seed-001-s1',  'awsS3',       'bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb', 'HEALTHY'),
  ('b0000001-0001-0000-0000-000000000003', 'a1b2c3d4-0001-0000-0000-000000000001', 0, 2, 'DATA',   'gdrive-seed-001-s2', 'googleDrive', 'cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc', 'HEALTHY'),
  ('b0000001-0001-0000-0000-000000000004', 'a1b2c3d4-0001-0000-0000-000000000001', 0, 3, 'DATA',   'awss3-seed-001-s3',  'awsS3',       'dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd', 'HEALTHY'),
  ('b0000001-0001-0000-0000-000000000005', 'a1b2c3d4-0001-0000-0000-000000000001', 0, 4, 'PARITY', 'gdrive-seed-001-s4', 'googleDrive', 'eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee', 'HEALTHY'),
  ('b0000001-0001-0000-0000-000000000006', 'a1b2c3d4-0001-0000-0000-000000000001', 0, 5, 'PARITY', 'awss3-seed-001-s5',  'awsS3',       'ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff', 'HEALTHY')
ON CONFLICT (file_id, chunk_index, shard_index) DO NOTHING;

-- database_export_2026.sql (UPLOADED — all 6 healthy)
INSERT INTO shards (id, file_id, chunk_index, shard_index, shard_type, remote_id, provider, checksum_sha256, status)
VALUES
  ('b0000002-0002-0000-0000-000000000001', 'a1b2c3d4-0002-0000-0000-000000000002', 0, 0, 'DATA',   'gdrive-seed-002-s0', 'googleDrive', 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa', 'HEALTHY'),
  ('b0000002-0002-0000-0000-000000000002', 'a1b2c3d4-0002-0000-0000-000000000002', 0, 1, 'DATA',   'awss3-seed-002-s1',  'awsS3',       'bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb', 'HEALTHY'),
  ('b0000002-0002-0000-0000-000000000003', 'a1b2c3d4-0002-0000-0000-000000000002', 0, 2, 'DATA',   'gdrive-seed-002-s2', 'googleDrive', 'cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc', 'HEALTHY'),
  ('b0000002-0002-0000-0000-000000000004', 'a1b2c3d4-0002-0000-0000-000000000002', 0, 3, 'DATA',   'awss3-seed-002-s3',  'awsS3',       'dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd', 'HEALTHY'),
  ('b0000002-0002-0000-0000-000000000005', 'a1b2c3d4-0002-0000-0000-000000000002', 0, 4, 'PARITY', 'gdrive-seed-002-s4', 'googleDrive', 'eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee', 'HEALTHY'),
  ('b0000002-0002-0000-0000-000000000006', 'a1b2c3d4-0002-0000-0000-000000000002', 0, 5, 'PARITY', 'awss3-seed-002-s5',  'awsS3',       'ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff', 'HEALTHY')
ON CONFLICT (file_id, chunk_index, shard_index) DO NOTHING;

-- sensitive_docs_encrypted.pdf (DEGRADED — shards 4-5 missing)
INSERT INTO shards (id, file_id, chunk_index, shard_index, shard_type, remote_id, provider, checksum_sha256, status)
VALUES
  ('b0000003-0003-0000-0000-000000000001', 'a1b2c3d4-0003-0000-0000-000000000003', 0, 0, 'DATA',   'gdrive-seed-003-s0', 'googleDrive', 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa', 'HEALTHY'),
  ('b0000003-0003-0000-0000-000000000002', 'a1b2c3d4-0003-0000-0000-000000000003', 0, 1, 'DATA',   'awss3-seed-003-s1',  'awsS3',       'bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb', 'HEALTHY'),
  ('b0000003-0003-0000-0000-000000000003', 'a1b2c3d4-0003-0000-0000-000000000003', 0, 2, 'DATA',   'gdrive-seed-003-s2', 'googleDrive', 'cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc', 'HEALTHY'),
  ('b0000003-0003-0000-0000-000000000004', 'a1b2c3d4-0003-0000-0000-000000000003', 0, 3, 'DATA',   'awss3-seed-003-s3',  'awsS3',       'dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd', 'HEALTHY'),
  ('b0000003-0003-0000-0000-000000000005', 'a1b2c3d4-0003-0000-0000-000000000003', 0, 4, 'PARITY', 'gdrive-seed-003-s4', 'googleDrive', 'eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee', 'MISSING'),
  ('b0000003-0003-0000-0000-000000000006', 'a1b2c3d4-0003-0000-0000-000000000003', 0, 5, 'PARITY', 'awss3-seed-003-s5',  'awsS3',       'ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff', 'MISSING')
ON CONFLICT (file_id, chunk_index, shard_index) DO NOTHING;

-- family_photos_2023.tar.gz (DEGRADED — shards 4-5 missing)
INSERT INTO shards (id, file_id, chunk_index, shard_index, shard_type, remote_id, provider, checksum_sha256, status)
VALUES
  ('b0000004-0004-0000-0000-000000000001', 'a1b2c3d4-0004-0000-0000-000000000004', 0, 0, 'DATA',   'gdrive-seed-004-s0', 'googleDrive', 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa', 'HEALTHY'),
  ('b0000004-0004-0000-0000-000000000002', 'a1b2c3d4-0004-0000-0000-000000000004', 0, 1, 'DATA',   'awss3-seed-004-s1',  'awsS3',       'bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb', 'HEALTHY'),
  ('b0000004-0004-0000-0000-000000000003', 'a1b2c3d4-0004-0000-0000-000000000004', 0, 2, 'DATA',   'gdrive-seed-004-s2', 'googleDrive', 'cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc', 'HEALTHY'),
  ('b0000004-0004-0000-0000-000000000004', 'a1b2c3d4-0004-0000-0000-000000000004', 0, 3, 'DATA',   'awss3-seed-004-s3',  'awsS3',       'dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd', 'HEALTHY'),
  ('b0000004-0004-0000-0000-000000000005', 'a1b2c3d4-0004-0000-0000-000000000004', 0, 4, 'PARITY', 'gdrive-seed-004-s4', 'googleDrive', 'eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee', 'MISSING'),
  ('b0000004-0004-0000-0000-000000000006', 'a1b2c3d4-0004-0000-0000-000000000004', 0, 5, 'PARITY', 'awss3-seed-004-s5',  'awsS3',       'ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff', 'MISSING')
ON CONFLICT (file_id, chunk_index, shard_index) DO NOTHING;

-- quarterly_report_Q4.xlsx (UPLOADED — all 6 healthy)
INSERT INTO shards (id, file_id, chunk_index, shard_index, shard_type, remote_id, provider, checksum_sha256, status)
VALUES
  ('b0000005-0005-0000-0000-000000000001', 'a1b2c3d4-0005-0000-0000-000000000005', 0, 0, 'DATA',   'gdrive-seed-005-s0', 'googleDrive', 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa', 'HEALTHY'),
  ('b0000005-0005-0000-0000-000000000002', 'a1b2c3d4-0005-0000-0000-000000000005', 0, 1, 'DATA',   'awss3-seed-005-s1',  'awsS3',       'bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb', 'HEALTHY'),
  ('b0000005-0005-0000-0000-000000000003', 'a1b2c3d4-0005-0000-0000-000000000005', 0, 2, 'DATA',   'gdrive-seed-005-s2', 'googleDrive', 'cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc', 'HEALTHY'),
  ('b0000005-0005-0000-0000-000000000004', 'a1b2c3d4-0005-0000-0000-000000000005', 0, 3, 'DATA',   'awss3-seed-005-s3',  'awsS3',       'dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd', 'HEALTHY'),
  ('b0000005-0005-0000-0000-000000000005', 'a1b2c3d4-0005-0000-0000-000000000005', 0, 4, 'PARITY', 'gdrive-seed-005-s4', 'googleDrive', 'eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee', 'HEALTHY'),
  ('b0000005-0005-0000-0000-000000000006', 'a1b2c3d4-0005-0000-0000-000000000005', 0, 5, 'PARITY', 'awss3-seed-005-s5',  'awsS3',       'ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff', 'HEALTHY')
ON CONFLICT (file_id, chunk_index, shard_index) DO NOTHING;
