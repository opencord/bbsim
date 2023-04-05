VOL-4883 - Publish a new bbsim release

https://github.com/opencord/bbsim/releases/tag/untagged-6d947de2a6078a1fe175
----------------------------------------------------------------------------
  o Draft bbsim release v2.96.8572 created by this patch set.
  o All 6cross-compile binaries are now included.
  o Change ID + commit message document the release.
  o Binaries downloaded/tested using {bbr,bbsim,bbsimctl} --help

Makefile
--------
  o Makefile release target is now able to cross compile all binaries:
    - bbsimctl cross compiled to arm, darwin, linux and windows.
    - release-{bbr, bbsim} only built *-linux-amd64
  o Refactor targets into a more modular and hierarchial setup.
  o Be sure to create RELEASE_DIR early on localhost so docker
    container can create a volume (NFS mount) for building into.
  o Display banners when building targets to improve log readability.
  o Clean targets updated to remove everything generated.
  o Help targets added to document indvidual target logic.
  o Makefile logic heavily commented.

makefiles/release/bbr.mk
makefiles/release/bbsim.mk
makefiles/release/bbsimctl.mk
-----------------------------
  o Relocate build-* and release-* target logic into named makefiles.
  o Update SHELL to set -euo pipefail to detect uninit shell vars.
  o Replaced shell/looping logic with true makefile target dependencies:
    - shell/loop replaced with a target rule that builds one platform binary.
    - Derive a list of targets release/{bbr,bbsim,bbsimctl}-{os}-{arch}
    - release: depends on all buildable targets
    - release rule can extract arch= and os= from target path being built.
    - To simplify logic construct env vars separately into a response file.
    - passed as "docker --env-file [x]" VS appending to an ever growing command line.
  o Marked more targets as .PHONY so they will always build.
  o Added helper target 'onf-publish' for exercising changes in the github-release script.

makefiles/include.mk
--------------------
  o Single include for importing all library makefiles.

makefiles/consts.mk
-------------------
  o Include fixed values, esp $(quote-single) which is now used to document
  o Refactor duplication in GO_* macros into named/reused values to shorten lines.

makefiles/lint
--------------
  o Moved lint::docker (hadolint target) into makefiles/lint/docker/*.mk.
  o Added makefiles/lint library targets for ability to lint more sources.

makefiles/tools.mk
------------------
  o Refactor DOCKER_* and GO_* macros into a named includable makefile.q