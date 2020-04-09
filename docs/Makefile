# Minimal makefile for Sphinx documentation

# use bash safe mode, fail quickly
SHELL = bash -e -o pipefail

# You can set these variables from the command line.
SPHINXOPTS    =
SPHINXBUILD   = sphinx-build
SOURCEDIR     = source
BUILDDIR      = build

.PHONY: help Makefile lint

# Put it first so that "make" without argument is like "make help".
help: doc_venv
	source $</bin/activate ; set -u ;\
	$(SPHINXBUILD) -M help "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

doc_venv:
	virtualenv -p python3 $@ ;\
	source $@/bin/activate ;\
	pip install -r requirements.txt

lint: doc8

doc8: doc_venv
	source $</bin/activate ; set -u ;\
	doc8 --max-line-length 119 \
	     $$(find . -name \*.rst ! -path "*doc_venv*")

clean:
	rm -rf doc_venv build

# Catch-all target: route all unknown targets to Sphinx using the new
# "make mode" option.  $(O) is meant as a shortcut for $(SPHINXOPTS).
%: doc_venv Makefile
	source $</bin/activate ; set -u ;\
	$(SPHINXBUILD) -M $@ "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

