package realip_zoning

import "iter"

func chan2Seq[T any](ch <-chan T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for item := range ch {
			if !yield(item) {
				return
			}
		}
	}
}

func seq2Chan[T any](seq iter.Seq[T]) <-chan T {
	pumper := func(seq iter.Seq[T], ch chan<- T) {
		defer close(ch)
		for item := range seq {
			ch <- item
		}
	}

	ch := make(chan T) // Unbuffered since we want it to be blocking
	go pumper(seq, ch)

	return ch
}

func mkMapper[U, T any](mapfunc func(T) U, input iter.Seq[T]) iter.Seq[U] {
	return func(yield func(U) bool) {
		for item := range input {
			if !yield(mapfunc(item)) {
				return
			}
		}
	}
}

type Zipped2[T01, T02 any] = struct {
	el01 T01
	el02 T02
}

func mkZip_short[A, B any](left iter.Seq[A], right iter.Seq[B]) iter.Seq[Zipped2[A, B]] {
	pulledA, stopA := iter.Pull(left)
	pulledB, stopB := iter.Pull(right)

	return func(yield func(Zipped2[A, B]) bool) {
		defer stopA()
		defer stopB()

		for {
			rightItem, rightOk := pulledB()
			leftItem, leftOk := pulledA()

			if !(leftOk && rightOk) {
				return
			}

			yieldWith := Zipped2[A, B]{
				el01: leftItem,
				el02: rightItem,
			}

			if !yield(yieldWith) {
				return
			}
		}
	}
}

func mkZip_long[A, B any](left iter.Seq[A], right iter.Seq[B]) iter.Seq[Zipped2[*A, *B]] {
	pulledA, stopA := iter.Pull(left)
	pulledB, stopB := iter.Pull(right)

	return func(yield func(Zipped2[*A, *B]) bool) {
		defer stopA()
		defer stopB()

		for {
			leftItem, leftOk := pulledA()
			rightItem, rightOk := pulledB()
			var ptrL *A
			var ptrR *B

			switch {
			case !leftOk && !rightOk:
				return
			case !leftOk:
				ptrL = nil
				ptrR = &rightItem
			case !rightOk:
				ptrL = &leftItem
				ptrR = nil
			}

			yieldWith := Zipped2[*A, *B]{
				el01: ptrL,
				el02: ptrR,
			}

			if !yield(yieldWith) {
				return
			}
		}
	}
}
