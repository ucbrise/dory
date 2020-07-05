#ifndef _COMMON_H
#define _COMMON_H

/*
 * Copyright (c) 2018, Henry Corrigan-Gibbs
 * 
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 * 
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
 * REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
 * FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
 * INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
 * LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
 * OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
 * PERFORMANCE OF THIS SOFTWARE.
 */



#include <stdbool.h>

/*
 * Return codes. For consistency with OpenSSL, we use 
 * non-zero values to denote success.
 */
#define OKAY 1
#define ERROR 0

/* Check a call that should return OKAY. */
#define CHECK_C(expr) do {\
  (rv = (expr));\
  if (rv != OKAY) {\
    goto cleanup;\
  }\
} while(false);

/* Check an allocation that should return non-NULL.*/
#define CHECK_A(expr) do {\
  (rv = ((expr) != NULL));\
  if (rv != OKAY) {\
    goto cleanup;\
  }\
} while(false);

#endif
