# CS464 Project TODO

### Application
- [ ] Refactor as monolith / Repackage as monolith
- [ ] Introduce concurrent deletion
- [ ] Optimize frontend (currently 150MB)
- [ ] Batch upload documents
- [ ] Add more storage providers (using SSH into NAS/local lab)
- [ ] Add more cloud providers (possibly automatically)
- [ ] Increase file size upload limit/provide explicit limit
<!-- - [ ] MAJOR: Develop script to automatically fetch cloud provider-specific secrets/tokens (agentic flow?)
- [ ] MAJOR: Develop cloud-native version
- [ ] MAJOR: Add smart provider selection strategy
- [ ] MAJOR: Introduce asynchronous deletion/uploading flow -->

### Reliability, Observability & Monitoring
- [ ] Introduce robust logging & self-troubleshooting
- [ ] Performance testing if possible
- [ ] Local resource optimization/monitoring if possible
<!-- - [ ] Add backup and restore workflows for SQLite databases
- [ ] Browser E2E tests -->


### Docs
- [ ] Add developer-oriented documentation to play around with the app
- [ ] Add steps to retrieve cloud provider-specific secrets/tokens

### DevOps
- [ ] Migrate from Dockerhub to Github Packages
- [ ] Introduce branch protection
- [ ] Trim branches
- [x] Rename microservices branch to dev branch