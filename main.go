package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// still bad names. but now the board is actual bitboards.
type m struct {
	a int
	b int
	c int
	d int
}

type bits struct {
	wp uint64
	wn uint64
	wb uint64
	wr uint64
	wq uint64
	wk uint64
	bp uint64
	bn uint64
	bb uint64
	br uint64
	bq uint64
	bk uint64
	ep int
	ca int
}

type oldd struct {
	old bits
}

type te struct {
	dep int
	val int
}

var boardForNoGoodReason bits
var junk int
var nodes int
var qnodes int
var stopTime time.Time
var stopNow bool
var tt map[string]te
var killer [80][2]m
var hist [8][8][8][8]int

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "uci" {
			uci()
			return
		}
		if os.Args[1] == "perft" {
			d := 3
			if len(os.Args) > 2 {
				n, err := strconv.Atoi(os.Args[2])
				if err == nil {
					d = n
				}
			}
			bo := st()
			fmt.Println(perft(bo, true, d))
			return
		}
	}

	boardForNoGoodReason = st()
	pb(boardForNoGoodReason)
	x := gen(boardForNoGoodReason, true)
	fmt.Println("white pseudo moves:", len(x))
	y := pick(boardForNoGoodReason, true)
	fmt.Println("beginner engine move:", mt(y))
	z := think(boardForNoGoodReason, true, 3, 500)
	fmt.Println("deeper engine move:", mt(z), "nodes:", nodes)
	fmt.Println("perft 2:", perft(boardForNoGoodReason, true, 2))
}

func st() bits {
	var z bits
	z.ep = -1
	z.ca = 15
	putp(&z, 0, 0, "r")
	putp(&z, 0, 1, "n")
	putp(&z, 0, 2, "b")
	putp(&z, 0, 3, "q")
	putp(&z, 0, 4, "k")
	putp(&z, 0, 5, "b")
	putp(&z, 0, 6, "n")
	putp(&z, 0, 7, "r")
	for i := 0; i < 8; i++ {
		putp(&z, 1, i, "p")
		putp(&z, 6, i, "P")
	}
	putp(&z, 7, 0, "R")
	putp(&z, 7, 1, "N")
	putp(&z, 7, 2, "B")
	putp(&z, 7, 3, "Q")
	putp(&z, 7, 4, "K")
	putp(&z, 7, 5, "B")
	putp(&z, 7, 6, "N")
	putp(&z, 7, 7, "R")
	return z
}

func empty() bits {
	return bits{ep: -1}
}

func pb(x bits) {
	for a := 0; a < 8; a++ {
		fmt.Print(8-a, " ")
		for b := 0; b < 8; b++ {
			fmt.Print(pc(x, a, b), " ")
		}
		fmt.Println("")
	}
	fmt.Println("  a b c d e f g h")
}

func sq(a int, b int) int {
	return a*8 + b
}

func bit(a int, b int) uint64 {
	if in(a, b) == false {
		return 0
	}
	return uint64(1) << uint(sq(a, b))
}

func bit2(s int) uint64 {
	if s < 0 || s > 63 {
		return 0
	}
	return uint64(1) << uint(s)
}

func first(x uint64) int {
	for i := 0; i < 64; i++ {
		if x&(uint64(1)<<uint(i)) != 0 {
			return i
		}
	}
	return -1
}

func pop(x uint64) int {
	n := 0
	for x != 0 {
		if x&1 == 1 {
			n++
		}
		x = x >> 1
	}
	return n
}

func wocc(x bits) uint64 {
	return x.wp | x.wn | x.wb | x.wr | x.wq | x.wk
}

func bocc(x bits) uint64 {
	return x.bp | x.bn | x.bb | x.br | x.bq | x.bk
}

func occ(x bits) uint64 {
	return wocc(x) | bocc(x)
}

func pc(x bits, a int, b int) string {
	mask := bit(a, b)
	if x.wp&mask != 0 {
		return "P"
	}
	if x.wn&mask != 0 {
		return "N"
	}
	if x.wb&mask != 0 {
		return "B"
	}
	if x.wr&mask != 0 {
		return "R"
	}
	if x.wq&mask != 0 {
		return "Q"
	}
	if x.wk&mask != 0 {
		return "K"
	}
	if x.bp&mask != 0 {
		return "p"
	}
	if x.bn&mask != 0 {
		return "n"
	}
	if x.bb&mask != 0 {
		return "b"
	}
	if x.br&mask != 0 {
		return "r"
	}
	if x.bq&mask != 0 {
		return "q"
	}
	if x.bk&mask != 0 {
		return "k"
	}
	return "."
}

func killat(x *bits, a int, b int) {
	mask := bit(a, b)
	x.wp = x.wp &^ mask
	x.wn = x.wn &^ mask
	x.wb = x.wb &^ mask
	x.wr = x.wr &^ mask
	x.wq = x.wq &^ mask
	x.wk = x.wk &^ mask
	x.bp = x.bp &^ mask
	x.bn = x.bn &^ mask
	x.bb = x.bb &^ mask
	x.br = x.br &^ mask
	x.bq = x.bq &^ mask
	x.bk = x.bk &^ mask
}

func putp(x *bits, a int, b int, p string) {
	mask := bit(a, b)
	killat(x, a, b)
	if p == "P" {
		x.wp = x.wp | mask
	}
	if p == "N" {
		x.wn = x.wn | mask
	}
	if p == "B" {
		x.wb = x.wb | mask
	}
	if p == "R" {
		x.wr = x.wr | mask
	}
	if p == "Q" {
		x.wq = x.wq | mask
	}
	if p == "K" {
		x.wk = x.wk | mask
	}
	if p == "p" {
		x.bp = x.bp | mask
	}
	if p == "n" {
		x.bn = x.bn | mask
	}
	if p == "b" {
		x.bb = x.bb | mask
	}
	if p == "r" {
		x.br = x.br | mask
	}
	if p == "q" {
		x.bq = x.bq | mask
	}
	if p == "k" {
		x.bk = x.bk | mask
	}
}

func gen(x bits, white bool) []m {
	var r []m
	if white {
		q := x.wp
		for q != 0 {
			s := first(q)
			q = q &^ bit2(s)
			a := s / 8
			b := s % 8
			if in(a-1, b) && occ(x)&bit(a-1, b) == 0 {
				r = append(r, m{a, b, a - 1, b})
				if a == 6 && occ(x)&bit(a-2, b) == 0 {
					r = append(r, m{a, b, a - 2, b})
				}
			}
			if in(a-1, b-1) && bocc(x)&bit(a-1, b-1) != 0 && pc(x, a-1, b-1) != "k" {
				r = append(r, m{a, b, a - 1, b - 1})
			}
			if in(a-1, b+1) && bocc(x)&bit(a-1, b+1) != 0 && pc(x, a-1, b+1) != "k" {
				r = append(r, m{a, b, a - 1, b + 1})
			}
			if x.ep >= 0 {
				ea := x.ep / 8
				eb := x.ep % 8
				if ea == a-1 && (eb == b-1 || eb == b+1) {
					r = append(r, m{a, b, ea, eb})
				}
			}
		}
		q = x.wn
		for q != 0 {
			s := first(q)
			q = q &^ bit2(s)
			r = horse(x, r, s/8, s%8, true)
		}
		q = x.wb
		for q != 0 {
			s := first(q)
			q = q &^ bit2(s)
			r = bish(x, r, s/8, s%8, true)
		}
		q = x.wr
		for q != 0 {
			s := first(q)
			q = q &^ bit2(s)
			r = rook(x, r, s/8, s%8, true)
		}
		q = x.wq
		for q != 0 {
			s := first(q)
			q = q &^ bit2(s)
			r = bish(x, r, s/8, s%8, true)
			r = rook(x, r, s/8, s%8, true)
		}
		q = x.wk
		if q != 0 {
			s := first(q)
			r = kingg(x, r, s/8, s%8, true)
			if pc(x, 7, 4) == "K" && ischk(x, true) == false {
				if x.ca&1 != 0 && occ(x)&(bit(7, 5)|bit(7, 6)) == 0 {
					if att(x, 7, 5, false) == false && att(x, 7, 6, false) == false {
						r = append(r, m{7, 4, 7, 6})
					}
				}
				if x.ca&2 != 0 && occ(x)&(bit(7, 1)|bit(7, 2)|bit(7, 3)) == 0 {
					if att(x, 7, 3, false) == false && att(x, 7, 2, false) == false {
						r = append(r, m{7, 4, 7, 2})
					}
				}
			}
		}
	} else {
		q := x.bp
		for q != 0 {
			s := first(q)
			q = q &^ bit2(s)
			a := s / 8
			b := s % 8
			if in(a+1, b) && occ(x)&bit(a+1, b) == 0 {
				r = append(r, m{a, b, a + 1, b})
				if a == 1 && occ(x)&bit(a+2, b) == 0 {
					r = append(r, m{a, b, a + 2, b})
				}
			}
			if in(a+1, b-1) && wocc(x)&bit(a+1, b-1) != 0 && pc(x, a+1, b-1) != "K" {
				r = append(r, m{a, b, a + 1, b - 1})
			}
			if in(a+1, b+1) && wocc(x)&bit(a+1, b+1) != 0 && pc(x, a+1, b+1) != "K" {
				r = append(r, m{a, b, a + 1, b + 1})
			}
			if x.ep >= 0 {
				ea := x.ep / 8
				eb := x.ep % 8
				if ea == a+1 && (eb == b-1 || eb == b+1) {
					r = append(r, m{a, b, ea, eb})
				}
			}
		}
		q = x.bn
		for q != 0 {
			s := first(q)
			q = q &^ bit2(s)
			r = horse(x, r, s/8, s%8, false)
		}
		q = x.bb
		for q != 0 {
			s := first(q)
			q = q &^ bit2(s)
			r = bish(x, r, s/8, s%8, false)
		}
		q = x.br
		for q != 0 {
			s := first(q)
			q = q &^ bit2(s)
			r = rook(x, r, s/8, s%8, false)
		}
		q = x.bq
		for q != 0 {
			s := first(q)
			q = q &^ bit2(s)
			r = bish(x, r, s/8, s%8, false)
			r = rook(x, r, s/8, s%8, false)
		}
		q = x.bk
		if q != 0 {
			s := first(q)
			r = kingg(x, r, s/8, s%8, false)
			if pc(x, 0, 4) == "k" && ischk(x, false) == false {
				if x.ca&4 != 0 && occ(x)&(bit(0, 5)|bit(0, 6)) == 0 {
					if att(x, 0, 5, true) == false && att(x, 0, 6, true) == false {
						r = append(r, m{0, 4, 0, 6})
					}
				}
				if x.ca&8 != 0 && occ(x)&(bit(0, 1)|bit(0, 2)|bit(0, 3)) == 0 {
					if att(x, 0, 3, true) == false && att(x, 0, 2, true) == false {
						r = append(r, m{0, 4, 0, 2})
					}
				}
			}
		}
	}
	return r
}

func horse(x bits, list []m, a int, b int, white bool) []m {
	list = tryy(x, list, a, b, a-2, b-1, white)
	list = tryy(x, list, a, b, a-2, b+1, white)
	list = tryy(x, list, a, b, a-1, b-2, white)
	list = tryy(x, list, a, b, a-1, b+2, white)
	list = tryy(x, list, a, b, a+1, b-2, white)
	list = tryy(x, list, a, b, a+1, b+2, white)
	list = tryy(x, list, a, b, a+2, b-1, white)
	list = tryy(x, list, a, b, a+2, b+1, white)
	return list
}

func bish(x bits, list []m, a int, b int, white bool) []m {
	list = slide(x, list, a, b, -1, -1, white)
	list = slide(x, list, a, b, -1, 1, white)
	list = slide(x, list, a, b, 1, -1, white)
	list = slide(x, list, a, b, 1, 1, white)
	return list
}

func rook(x bits, list []m, a int, b int, white bool) []m {
	list = slide(x, list, a, b, -1, 0, white)
	list = slide(x, list, a, b, 1, 0, white)
	list = slide(x, list, a, b, 0, -1, white)
	list = slide(x, list, a, b, 0, 1, white)
	return list
}

func kingg(x bits, list []m, a int, b int, white bool) []m {
	list = tryy(x, list, a, b, a-1, b-1, white)
	list = tryy(x, list, a, b, a-1, b, white)
	list = tryy(x, list, a, b, a-1, b+1, white)
	list = tryy(x, list, a, b, a, b-1, white)
	list = tryy(x, list, a, b, a, b+1, white)
	list = tryy(x, list, a, b, a+1, b-1, white)
	list = tryy(x, list, a, b, a+1, b, white)
	list = tryy(x, list, a, b, a+1, b+1, white)
	return list
}

func slide(x bits, list []m, a int, b int, aa int, bb int, white bool) []m {
	q := a + aa
	e := b + bb
	for in(q, e) {
		p := pc(x, q, e)
		if p == "." {
			list = append(list, m{a, b, q, e})
		} else {
			if p == "K" || p == "k" {
				return list
			}
			if white && bocc(x)&bit(q, e) != 0 {
				list = append(list, m{a, b, q, e})
			}
			if white == false && wocc(x)&bit(q, e) != 0 {
				list = append(list, m{a, b, q, e})
			}
			return list
		}
		q = q + aa
		e = e + bb
	}
	return list
}

func tryy(x bits, list []m, a int, b int, c int, d int, white bool) []m {
	if in(c, d) == false {
		return list
	}
	p := pc(x, c, d)
	if p == "." {
		list = append(list, m{a, b, c, d})
		return list
	}
	if p == "K" || p == "k" {
		return list
	}
	if white && bocc(x)&bit(c, d) != 0 {
		list = append(list, m{a, b, c, d})
	}
	if white == false && wocc(x)&bit(c, d) != 0 {
		list = append(list, m{a, b, c, d})
	}
	return list
}

func leg(x bits, white bool) []m {
	z := gen(x, white)
	var good []m
	for i := 0; i < len(z); i++ {
		y := x
		domove(&y, z[i])
		if ischk(y, white) == false {
			good = append(good, z[i])
		}
	}
	return good
}

func legp(x *bits, white bool) []m {
	z := gen(*x, white)
	var good []m
	for i := 0; i < len(z); i++ {
		u := dobad(x, z[i])
		if ischk(*x, white) == false {
			good = append(good, z[i])
		}
		undobad(x, u)
	}
	return good
}

func dobad(x *bits, y m) oldd {
	o := oldd{old: *x}
	domove(x, y)
	return o
}

func undobad(x *bits, o oldd) {
	*x = o.old
}

func domove(x *bits, y m) {
	p := pc(*x, y.a, y.b)
	cap := pc(*x, y.c, y.d)
	x.ep = -1

	if p == "K" {
		x.ca = x.ca &^ 3
	}
	if p == "k" {
		x.ca = x.ca &^ 12
	}
	if p == "R" && y.a == 7 && y.b == 0 {
		x.ca = x.ca &^ 2
	}
	if p == "R" && y.a == 7 && y.b == 7 {
		x.ca = x.ca &^ 1
	}
	if p == "r" && y.a == 0 && y.b == 0 {
		x.ca = x.ca &^ 8
	}
	if p == "r" && y.a == 0 && y.b == 7 {
		x.ca = x.ca &^ 4
	}
	if cap == "R" && y.c == 7 && y.d == 0 {
		x.ca = x.ca &^ 2
	}
	if cap == "R" && y.c == 7 && y.d == 7 {
		x.ca = x.ca &^ 1
	}
	if cap == "r" && y.c == 0 && y.d == 0 {
		x.ca = x.ca &^ 8
	}
	if cap == "r" && y.c == 0 && y.d == 7 {
		x.ca = x.ca &^ 4
	}

	oldEp := x.ep
	_ = oldEp
	if p == "P" && y.a == 6 && y.c == 4 {
		x.ep = sq(5, y.b)
	}
	if p == "p" && y.a == 1 && y.c == 3 {
		x.ep = sq(2, y.b)
	}

	if p == "P" && cap == "." && y.b != y.d {
		killat(x, y.c+1, y.d)
	}
	if p == "p" && cap == "." && y.b != y.d {
		killat(x, y.c-1, y.d)
	}

	killat(x, y.a, y.b)
	killat(x, y.c, y.d)

	if p == "P" && y.c == 0 {
		putp(x, y.c, y.d, "Q")
	} else if p == "p" && y.c == 7 {
		putp(x, y.c, y.d, "q")
	} else {
		putp(x, y.c, y.d, p)
	}

	if p == "K" && y.a == 7 && y.b == 4 && y.d == 6 {
		killat(x, 7, 7)
		putp(x, 7, 5, "R")
	}
	if p == "K" && y.a == 7 && y.b == 4 && y.d == 2 {
		killat(x, 7, 0)
		putp(x, 7, 3, "R")
	}
	if p == "k" && y.a == 0 && y.b == 4 && y.d == 6 {
		killat(x, 0, 7)
		putp(x, 0, 5, "r")
	}
	if p == "k" && y.a == 0 && y.b == 4 && y.d == 2 {
		killat(x, 0, 0)
		putp(x, 0, 3, "r")
	}
}

func ischk(x bits, white bool) bool {
	k := x.wk
	if white == false {
		k = x.bk
	}
	s := first(k)
	if s < 0 {
		return true
	}
	return att(x, s/8, s%8, !white)
}

func att(x bits, a int, b int, bywhite bool) bool {
	if bywhite {
		if in(a+1, b-1) && x.wp&bit(a+1, b-1) != 0 {
			return true
		}
		if in(a+1, b+1) && x.wp&bit(a+1, b+1) != 0 {
			return true
		}
	} else {
		if in(a-1, b-1) && x.bp&bit(a-1, b-1) != 0 {
			return true
		}
		if in(a-1, b+1) && x.bp&bit(a-1, b+1) != 0 {
			return true
		}
	}
	nn := x.wn
	kk := x.wk
	bp := "B"
	rp := "R"
	qp := "Q"
	if bywhite == false {
		nn = x.bn
		kk = x.bk
		bp = "b"
		rp = "r"
		qp = "q"
	}
	if in(a-2, b-1) && nn&bit(a-2, b-1) != 0 {
		return true
	}
	if in(a-2, b+1) && nn&bit(a-2, b+1) != 0 {
		return true
	}
	if in(a-1, b-2) && nn&bit(a-1, b-2) != 0 {
		return true
	}
	if in(a-1, b+2) && nn&bit(a-1, b+2) != 0 {
		return true
	}
	if in(a+1, b-2) && nn&bit(a+1, b-2) != 0 {
		return true
	}
	if in(a+1, b+2) && nn&bit(a+1, b+2) != 0 {
		return true
	}
	if in(a+2, b-1) && nn&bit(a+2, b-1) != 0 {
		return true
	}
	if in(a+2, b+1) && nn&bit(a+2, b+1) != 0 {
		return true
	}
	for aa := -1; aa <= 1; aa++ {
		for bb := -1; bb <= 1; bb++ {
			if aa == 0 && bb == 0 {
				junk = junk + 0
			} else if in(a+aa, b+bb) && kk&bit(a+aa, b+bb) != 0 {
				return true
			}
		}
	}
	if ray(x, a, b, -1, -1, bp, qp) {
		return true
	}
	if ray(x, a, b, -1, 1, bp, qp) {
		return true
	}
	if ray(x, a, b, 1, -1, bp, qp) {
		return true
	}
	if ray(x, a, b, 1, 1, bp, qp) {
		return true
	}
	if ray(x, a, b, -1, 0, rp, qp) {
		return true
	}
	if ray(x, a, b, 1, 0, rp, qp) {
		return true
	}
	if ray(x, a, b, 0, -1, rp, qp) {
		return true
	}
	if ray(x, a, b, 0, 1, rp, qp) {
		return true
	}
	return false
}

func ray(x bits, a int, b int, aa int, bb int, p1 string, p2 string) bool {
	r := a + aa
	c := b + bb
	for in(r, c) {
		p := pc(x, r, c)
		if p != "." {
			if p == p1 || p == p2 {
				return true
			}
			return false
		}
		r = r + aa
		c = c + bb
	}
	return false
}

func pick(x bits, white bool) m {
	l := gen(x, white)
	if len(l) == 0 {
		return m{-1, -1, -1, -1}
	}
	best := l[0]
	score := -999999
	if white == false {
		score = 999999
	}
	for i := 0; i < len(l); i++ {
		y := x
		domove(&y, l[i])
		s := mat(y)
		if white && s > score {
			score = s
			best = l[i]
		}
		if white == false && s < score {
			score = s
			best = l[i]
		}
	}
	return best
}

func mat(x bits) int {
	return pop(x.wp)*100 + pop(x.wn)*320 + pop(x.wb)*330 + pop(x.wr)*500 + pop(x.wq)*900 -
		pop(x.bp)*100 - pop(x.bn)*320 - pop(x.bb)*330 - pop(x.br)*500 - pop(x.bq)*900
}

func val(p string) int {
	if p == "P" || p == "p" {
		return 100
	}
	if p == "N" || p == "n" {
		return 320
	}
	if p == "B" || p == "b" {
		return 330
	}
	if p == "R" || p == "r" {
		return 500
	}
	if p == "Q" || p == "q" {
		return 900
	}
	if p == "K" || p == "k" {
		return 20000
	}
	return 0
}

func ev(x bits) int {
	s := mat(x)
	for a := 0; a < 8; a++ {
		for b := 0; b < 8; b++ {
			p := pc(x, a, b)
			if p == "P" {
				s = s + (6-a)*4
				if b == 3 || b == 4 {
					s = s + 7
				}
			}
			if p == "p" {
				s = s - (a-1)*4
				if b == 3 || b == 4 {
					s = s - 7
				}
			}
			if p == "N" && a >= 2 && a <= 5 && b >= 2 && b <= 5 {
				s = s + 20
			}
			if p == "n" && a >= 2 && a <= 5 && b >= 2 && b <= 5 {
				s = s - 20
			}
			if p == "K" && a == 7 && (b == 6 || b == 2) {
				s = s + 15
			}
			if p == "k" && a == 0 && (b == 6 || b == 2) {
				s = s - 15
			}
		}
	}
	s = s + len(gen(x, true))*2
	s = s - len(gen(x, false))*2
	if ischk(x, false) {
		s = s + 30
	}
	if ischk(x, true) {
		s = s - 30
	}
	return s
}

func perft(x bits, white bool, dep int) int64 {
	if dep == 0 {
		return 1
	}
	l := leg(x, white)
	if dep == 1 {
		return int64(len(l))
	}
	var total int64
	for i := 0; i < len(l); i++ {
		y := x
		domove(&y, l[i])
		total = total + perft(y, !white, dep-1)
	}
	return total
}

func perftu(x *bits, white bool, dep int) int64 {
	if dep == 0 {
		return 1
	}
	l := legp(x, white)
	if dep == 1 {
		return int64(len(l))
	}
	var total int64
	for i := 0; i < len(l); i++ {
		o := dobad(x, l[i])
		total = total + perftu(x, !white, dep-1)
		undobad(x, o)
	}
	return total
}

func mini(x bits, white bool, dep int) int {
	nodes = nodes + 1
	if dep == 0 {
		return ev(x)
	}
	l := leg(x, white)
	if len(l) == 0 {
		if ischk(x, white) {
			if white {
				return -900000
			}
			return 900000
		}
		return 0
	}
	if white {
		best := -9999999
		for i := 0; i < len(l); i++ {
			y := x
			domove(&y, l[i])
			v := mini(y, false, dep-1)
			if v > best {
				best = v
			}
		}
		return best
	}
	best := 9999999
	for i := 0; i < len(l); i++ {
		y := x
		domove(&y, l[i])
		v := mini(y, true, dep-1)
		if v < best {
			best = v
		}
	}
	return best
}

func rootmini(x bits, white bool, dep int) m {
	l := leg(x, white)
	if len(l) == 0 {
		return m{-1, -1, -1, -1}
	}
	best := l[0]
	if white {
		bs := -9999999
		for i := 0; i < len(l); i++ {
			y := x
			domove(&y, l[i])
			v := mini(y, false, dep-1)
			if v > bs {
				bs = v
				best = l[i]
			}
		}
	} else {
		bs := 9999999
		for i := 0; i < len(l); i++ {
			y := x
			domove(&y, l[i])
			v := mini(y, true, dep-1)
			if v < bs {
				bs = v
				best = l[i]
			}
		}
	}
	return best
}

func capsonly(x *bits, white bool) []m {
	l := legp(x, white)
	var r []m
	for i := 0; i < len(l); i++ {
		if pc(*x, l[i].c, l[i].d) != "." || sq(l[i].c, l[i].d) == x.ep {
			r = append(r, l[i])
		}
	}
	return r
}

func qs(x *bits, white bool, alpha int, beta int) int {
	qnodes = qnodes + 1
	stand := ev(*x)
	if white {
		if stand >= beta {
			return beta
		}
		if stand > alpha {
			alpha = stand
		}
		l := capsonly(x, white)
		l = ord(*x, l, 0)
		for i := 0; i < len(l); i++ {
			o := dobad(x, l[i])
			v := qs(x, false, alpha, beta)
			undobad(x, o)
			if v >= beta {
				return beta
			}
			if v > alpha {
				alpha = v
			}
		}
		return alpha
	}
	if stand <= alpha {
		return alpha
	}
	if stand < beta {
		beta = stand
	}
	l := capsonly(x, white)
	l = ord(*x, l, 0)
	for i := 0; i < len(l); i++ {
		o := dobad(x, l[i])
		v := qs(x, true, alpha, beta)
		undobad(x, o)
		if v <= alpha {
			return alpha
		}
		if v < beta {
			beta = v
		}
	}
	return beta
}

func abfast(x *bits, white bool, dep int, alpha int, beta int) int {
	nodes = nodes + 1
	if stopTime.IsZero() == false && time.Now().After(stopTime) {
		stopNow = true
		return ev(*x)
	}
	if dep == 0 {
		return qs(x, white, alpha, beta)
	}
	if tt == nil {
		tt = make(map[string]te)
	}
	k := hhh(*x, white, dep)
	got, ok := tt[k]
	if ok && got.dep >= dep {
		return got.val
	}
	if dep >= 3 && ischk(*x, white) == false {
		if white {
			nv := abfast(x, false, dep-3, alpha, beta)
			if nv >= beta {
				return beta
			}
		} else {
			nv := abfast(x, true, dep-3, alpha, beta)
			if nv <= alpha {
				return alpha
			}
		}
	}
	l := legp(x, white)
	l = ord(*x, l, dep)
	if len(l) == 0 {
		if ischk(*x, white) {
			if white {
				return -900000 - dep
			}
			return 900000 + dep
		}
		return 0
	}
	ans := 0
	if white {
		ans = -9999999
		for i := 0; i < len(l); i++ {
			o := dobad(x, l[i])
			v := abfast(x, false, dep-1, alpha, beta)
			undobad(x, o)
			if v > ans {
				ans = v
			}
			if ans > alpha {
				alpha = ans
			}
			if beta <= alpha {
				saveKill(dep, l[i])
				break
			}
		}
	} else {
		ans = 9999999
		for i := 0; i < len(l); i++ {
			o := dobad(x, l[i])
			v := abfast(x, true, dep-1, alpha, beta)
			undobad(x, o)
			if v < ans {
				ans = v
			}
			if ans < beta {
				beta = ans
			}
			if beta <= alpha {
				saveKill(dep, l[i])
				break
			}
		}
	}
	if len(tt) < 200000 {
		tt[k] = te{dep: dep, val: ans}
	}
	return ans
}

func rootab(x *bits, white bool, dep int) m {
	l := legp(x, white)
	l = ord(*x, l, dep)
	if len(l) == 0 {
		return m{-1, -1, -1, -1}
	}
	best := l[0]
	if white {
		bs := -9999999
		for i := 0; i < len(l); i++ {
			o := dobad(x, l[i])
			v := abfast(x, false, dep-1, -9999999, 9999999)
			undobad(x, o)
			if stopNow {
				return best
			}
			if v > bs {
				bs = v
				best = l[i]
			}
		}
	} else {
		bs := 9999999
		for i := 0; i < len(l); i++ {
			o := dobad(x, l[i])
			v := abfast(x, true, dep-1, -9999999, 9999999)
			undobad(x, o)
			if stopNow {
				return best
			}
			if v < bs {
				bs = v
				best = l[i]
			}
		}
	}
	return best
}

func think(x bits, white bool, dep int, ms int) m {
	nodes = 0
	qnodes = 0
	stopNow = false
	tt = make(map[string]te)
	if ms > 0 {
		stopTime = time.Now().Add(time.Duration(ms) * time.Millisecond)
	} else {
		stopTime = time.Time{}
	}
	best := m{-1, -1, -1, -1}
	for d := 1; d <= dep; d++ {
		y := x
		v := rootab(&y, white, d)
		if stopNow == false {
			best = v
		}
		if stopNow {
			break
		}
	}
	stopTime = time.Time{}
	return best
}

func ord(x bits, l []m, dep int) []m {
	for i := 0; i < len(l); i++ {
		for j := i + 1; j < len(l); j++ {
			if scmove(x, l[j], dep) > scmove(x, l[i], dep) {
				tmp := l[i]
				l[i] = l[j]
				l[j] = tmp
			}
		}
	}
	return l
}

func scmove(x bits, y m, dep int) int {
	s := 0
	if pc(x, y.c, y.d) != "." {
		s = s + val(pc(x, y.c, y.d))*10 - val(pc(x, y.a, y.b))
	}
	if dep >= 0 && dep < 80 {
		if same(y, killer[dep][0]) {
			s = s + 9000
		}
		if same(y, killer[dep][1]) {
			s = s + 8000
		}
	}
	s = s + hist[y.a][y.b][y.c][y.d]
	return s
}

func same(a m, b m) bool {
	return a.a == b.a && a.b == b.b && a.c == b.c && a.d == b.d
}

func saveKill(dep int, y m) {
	if dep < 0 || dep >= 80 {
		return
	}
	if same(killer[dep][0], y) == false {
		killer[dep][1] = killer[dep][0]
		killer[dep][0] = y
	}
	hist[y.a][y.b][y.c][y.d] = hist[y.a][y.b][y.c][y.d] + dep*dep
}

func hhh(x bits, white bool, dep int) string {
	return strconv.FormatUint(x.wp, 16) + "/" + strconv.FormatUint(x.wn, 16) + "/" +
		strconv.FormatUint(x.wb, 16) + "/" + strconv.FormatUint(x.wr, 16) + "/" +
		strconv.FormatUint(x.wq, 16) + "/" + strconv.FormatUint(x.wk, 16) + "/" +
		strconv.FormatUint(x.bp, 16) + "/" + strconv.FormatUint(x.bn, 16) + "/" +
		strconv.FormatUint(x.bb, 16) + "/" + strconv.FormatUint(x.br, 16) + "/" +
		strconv.FormatUint(x.bq, 16) + "/" + strconv.FormatUint(x.bk, 16) + "/" +
		strconv.Itoa(x.ep) + "/" + strconv.Itoa(x.ca) + "/" + fmt.Sprint(white) + "/" + strconv.Itoa(dep)
}

func bitcount(x bits) int {
	return pop(occ(x))
}

func fen(s string) (bits, bool) {
	x := empty()
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return st(), true
	}
	r := 0
	c := 0
	for i := 0; i < len(parts[0]); i++ {
		ch := parts[0][i]
		if ch == '/' {
			r = r + 1
			c = 0
		} else if ch >= '1' && ch <= '8' {
			c = c + int(ch-'0')
		} else {
			if in(r, c) {
				putp(&x, r, c, string(ch))
			}
			c = c + 1
		}
	}
	wturn := true
	if len(parts) > 1 && parts[1] == "b" {
		wturn = false
	}
	x.ca = 0
	if len(parts) > 2 {
		if strings.Contains(parts[2], "K") {
			x.ca = x.ca | 1
		}
		if strings.Contains(parts[2], "Q") {
			x.ca = x.ca | 2
		}
		if strings.Contains(parts[2], "k") {
			x.ca = x.ca | 4
		}
		if strings.Contains(parts[2], "q") {
			x.ca = x.ca | 8
		}
	}
	x.ep = -1
	if len(parts) > 3 && parts[3] != "-" {
		mv := um(parts[3] + parts[3])
		x.ep = sq(mv.a, mv.b)
	}
	return x, wturn
}

func um(s string) m {
	if len(s) < 4 {
		return m{-1, -1, -1, -1}
	}
	a := 8 - int(s[1]-'0')
	b := int(s[0] - 'a')
	c := 8 - int(s[3]-'0')
	d := int(s[2] - 'a')
	return m{a, b, c, d}
}

func playtxt(x *bits, white bool, s string) {
	y := um(s)
	l := legp(x, white)
	for i := 0; i < len(l); i++ {
		if same(l[i], y) {
			domove(x, l[i])
			return
		}
	}
	if in(y.a, y.b) && in(y.c, y.d) {
		domove(x, y)
	}
}

func uci() {
	inp := bufio.NewScanner(os.Stdin)
	bo := st()
	white := true
	for inp.Scan() {
		line := strings.TrimSpace(inp.Text())
		if line == "uci" {
			fmt.Println("id name chesseg-bitboard-bad-beginner")
			fmt.Println("id author beginner-ish")
			fmt.Println("uciok")
		} else if line == "isready" {
			fmt.Println("readyok")
		} else if strings.HasPrefix(line, "position") {
			bo, white = posline(line)
		} else if strings.HasPrefix(line, "go") {
			dep := 4
			ms := 1000
			p := strings.Fields(line)
			for i := 0; i < len(p); i++ {
				if p[i] == "depth" && i+1 < len(p) {
					n, e := strconv.Atoi(p[i+1])
					if e == nil {
						dep = n
						ms = 0
					}
				}
				if p[i] == "movetime" && i+1 < len(p) {
					n, e := strconv.Atoi(p[i+1])
					if e == nil {
						ms = n
					}
				}
			}
			mv := think(bo, white, dep, ms)
			fmt.Println("bestmove", mt(mv))
		} else if line == "ucinewgame" {
			bo = st()
			white = true
			tt = make(map[string]te)
		} else if line == "quit" {
			return
		}
	}
}

func posline(line string) (bits, bool) {
	parts := strings.Fields(line)
	bo := st()
	white := true
	i := 1
	if len(parts) <= 1 {
		return bo, white
	}
	if parts[i] == "startpos" {
		bo = st()
		white = true
		i = i + 1
	} else if parts[i] == "fen" {
		i = i + 1
		var fparts []string
		for i < len(parts) && parts[i] != "moves" {
			fparts = append(fparts, parts[i])
			i = i + 1
		}
		bo, white = fen(strings.Join(fparts, " "))
	}
	if i < len(parts) && parts[i] == "moves" {
		i = i + 1
		for i < len(parts) {
			playtxt(&bo, white, parts[i])
			white = !white
			i = i + 1
		}
	}
	return bo, white
}

func selfgame(ply int) int {
	bo := st()
	white := true
	made := 0
	for i := 0; i < ply; i++ {
		mv := think(bo, white, 2, 0)
		if mv.a == -1 {
			break
		}
		l := leg(bo, white)
		ok := false
		for j := 0; j < len(l); j++ {
			if same(mv, l[j]) {
				ok = true
			}
		}
		if ok == false {
			break
		}
		domove(&bo, mv)
		white = !white
		made = made + 1
	}
	return made
}

func in(a int, b int) bool {
	if a < 0 {
		return false
	}
	if a > 7 {
		return false
	}
	if b < 0 {
		return false
	}
	if b > 7 {
		return false
	}
	return true
}

func mt(x m) string {
	if x.a == -1 {
		return "no move"
	}
	f := "abcdefgh"
	return string(f[x.b]) + fmt.Sprint(8-x.a) + string(f[x.d]) + fmt.Sprint(8-x.c)
}
