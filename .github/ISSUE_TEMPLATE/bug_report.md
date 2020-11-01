---
name: Bug Report
about: Report a bug encountered while using Local PVs
labels: kind/bug

---

<!-- Please use this template while reporting a bug and provide as much info as possible.
     Not doing so may result in your bug not being addressed in a timely manner. Thanks!
-->

**Describe the bug:** A clear and concise description of what the bug is.

**Expected behaviour:** A concise description of what you expected to happen

**Steps to reproduce the bug:**
Steps to reproduce the bug should be clear and easily reproducible to help people gain an understanding of the problem

**The output of the following commands will help us better understand what's going on**:
<!-- (Pasting long output into a [GitHub gist](https://gist.github.com) or other [Pastebin](https://pastebin.com/) is fine.) -->

* `kubectl get pods -n <openebs_namespace> --show-labels`
* `kubectl logs <upgrade_job_pod> -n <openebs_namespace>`

**Anything else we need to know?:**
Add any other context about the problem here.

**Environment details:**
- OpenEBS version (use `kubectl get po -n openebs --show-labels`):
- Kubernetes version (use `kubectl version`):
- Cloud provider or hardware configuration:
- OS (e.g: `cat /etc/os-release`):
- kernel (e.g: `uname -a`):
- others:
