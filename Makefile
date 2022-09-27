.PHONY: doc
doc:
	@echo "Available targets:"
	@cat Makefile | grep ^.PHONY | sed 's/^.PHONY:/ /g'

.PHONY: buildcontainer
buildcontainer:
	./scripts/buildc.bash

.PHONY: runcontainer
runcontainer:
	./scripts/runc.bash
