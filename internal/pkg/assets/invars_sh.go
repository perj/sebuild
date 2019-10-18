// Copyright 2018-2019 Schibsted

package assets

const InvarsSh = `
set -e

if [ "$(command -v setval)" != "setval" ]; then
	setval() {
		local _k=$1
		shift
		eval "$_k=\"$*\""
		echo "$_k=$*"
	}
fi
if [ "$(command -v depend)" != "depend" ]; then
	if [ -n "$depfile" ]; then
		depend() {
			for dep in $@; do
				echo -n " ${dep}" >> ${depfile}
			done
		}
	else
		depend() {
			:
		}
	fi
fi

setval BUILD_STAGE $(echo $buildflavor | tr [:lower:] [:upper:])

setval GOARCH $(go env GOARCH 2>/dev/null)
setval GOOS $(go env GOOS 2>/dev/null)
`
