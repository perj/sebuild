// Copyright 2018 Schibsted

#include <stdio.h>
#include <stdlib.h>

#include "sections.h"

int
main(int argc, char *argv[]) {
	if (argc != 2) {
		fprintf(stderr, "Usage: %s <number>\n", argv[0]);
		fprintf(stderr, "Prefix the number with 0x for hexadecimal or just 0 for octal.\n");
		return 1;
	}

	char *number = argv[1];
	char *end = NULL;

	long uch = strtol(number, &end, 0);
	if (end == number || *end != '\0') {
		fprintf(stderr, "Failed to parse number.\n");
		return 1;
	}
	if (uch < 0 || uch > 0x10FFFF) {
		fprintf(stderr, "Number is out of range for Unicode codepoints.\n");
		return 1;
	}

	// Could binary search, but linear is also quick.
	struct section *s;
	for (s = sections ; s->name ; s++) {
		if (uch >= s->start && uch <= s->end)
			break;
	}

	if (!s->name) {
		fprintf(stderr, "Failed to find section name for codepoint U+%lX.\n", uch);
		return 1;
	}

	printf("%s\n", s->name);
	return 0;
}
