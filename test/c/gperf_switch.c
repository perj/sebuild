// Copyright 2019 Schibsted

#include "build_ctest.h"

#include <stdio.h>
#include <stdlib.h>

#include "foo.h"

void
gperf_enum_test(void) {
	GPERF_ENUM(foo)
	switch (lookup_foo("bar", -1)) {
	case GPERF_CASE("foo"):
		printf("got foo\n");
		exit(1);
	case GPERF_CASE("bar"):
		break;
	case GPERF_CASE_NONE:
		printf("got none\n");
		exit(1);
	}
}
