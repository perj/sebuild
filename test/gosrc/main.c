// Copyright 2019 Schibsted

#include "gosrc.h"

void init_go_runtime(int argc, char *const *argv);

int
main(int argc, char *argv[]) {
	init_go_runtime(argc, argv);
	GosrcTest();
}
