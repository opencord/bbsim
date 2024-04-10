bbsim: TODO
===========

- Relocate build/package/Dockerfile into a different directory.
- Perhaps dockerfile/* ?
- Directory build/ is commonly used for:
  - Storing generated content.
  - An output directory for build system jobs.
  - Directory build/ is transient, removed and recreated between builds.
- Placing revision control files beneath build/ will create
  problems for '% make clean' and cleanup scripts.
