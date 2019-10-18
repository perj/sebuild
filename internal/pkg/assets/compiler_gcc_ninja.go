// Copyright 2018-2019 Schibsted

package assets

const CompilerGccNinja = `
# clang does not support -Wlogical-op so only set it for gcc.
warncompiler=-Wlogical-op
`
