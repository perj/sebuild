// Copyright 2019 Schibsted

#include "gosrc.h"

#ifdef INIT_GO
void init_go_runtime(int argc, char *const *argv);
#endif

int
main(int argc, char *argv[]) {
#ifdef INIT_GO
	init_go_runtime(argc, argv);
#endif
	GosrcTest();
}
