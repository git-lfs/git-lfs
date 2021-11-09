/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package plurals

type math interface {
	calc(n uint32) uint32
}

type mod struct {
	value uint32
}

func (m mod) calc(n uint32) uint32 {
	return n % m.value
}
