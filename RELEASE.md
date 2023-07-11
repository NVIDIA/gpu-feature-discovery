# Release Process

The `gpu-feature-discovery` component consists in two artifacts:
- The `gpu-feature-discovery` container
- The `gpu-feature-discovery` helm chart

Publishing the container is automated through gitlab-ci and only requires one to tag the commit and push it to gitlab.
Publishing the helm chart is currently manual, and we should move to an automated process ASAP

# Release Process Checklist
- [ ] Update the README to change occurances of the old version (e.g: `v0.8.1`) with the new version
- [ ] Commit, Tag and Push to Gitlab
- [ ] Build a new helm package with `helm package ./deployments/helm/gpu-feature-discovery`
- [ ] Switch to the `gh-pages` branch and move the newly generated package to either the `stable` helm repo
- [ ] Run the `./build-index.sh` script to rebuild the indices for each repo
- [ ] Commit and push the `gh-pages` branch to GitHub
- [ ] Wait for the [CI job associated with your tag] (https://gitlab.com/nvidia/kubernetes/gpu-feature-discovery/-/pipelines) to complete
- [ ] Create a [new release](https://github.com/NVIDIA/gpu-feature-discovery/releases) on Github with the changelog
