TOOLS_BINDIR = $(realpath $(TOOLS_DIR)/bin)
TKN_VERSION = 0.31.1

.PHONY: install-out-of-tree-tools
install-out-of-tree-tools: \
	$(TOOLS_BINDIR)/tkn 

$(TOOLS_BINDIR)/tkn: 
	cd $(TOOLS_BINDIR) \
	&& curl -LO "https://github.com/tektoncd/cli/releases/download/v${TKN_VERSION}/tkn_${TKN_VERSION}_Linux_x86_64.tar.gz" \
	&& tar xvzf "tkn_${TKN_VERSION}_Linux_x86_64.tar.gz" tkn \
	&& rm "tkn_${TKN_VERSION}_Linux_x86_64.tar.gz"
