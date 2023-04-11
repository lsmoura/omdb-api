AUTH_TOKEN ?= TODO

vercel.json: vercel.base.json
	@sed -e 's/$$AUTH_TOKEN/$(AUTH_TOKEN)/' $< > $@
