package main

import "testing"

func TestStartingPositionHasTwentyWhiteMoves(t *testing.T) {
	b := st()
	moves := gen(b, true)
	if len(moves) != 20 {
		t.Fatalf("wanted 20 moves, got %d", len(moves))
	}
}

func TestBlackHasTwentyMovesAfterE2E4(t *testing.T) {
	bo := st()
	domove(&bo, m{6, 4, 4, 4})
	thing := gen(bo, false)
	if len(thing) != 20 {
		t.Fatalf("wanted 20 moves, got %d", len(thing))
	}
}

func TestMaterialStartsEqual(t *testing.T) {
	bbb := st()
	if mat(bbb) != 0 {
		t.Fatalf("wanted material score 0, got %d", mat(bbb))
	}
}

func TestMoveText(t *testing.T) {
	if mt(m{6, 4, 4, 4}) != "e2e4" {
		t.Fatalf("wanted e2e4, got %s", mt(m{6, 4, 4, 4}))
	}
}

func TestLegalStartStillTwenty(t *testing.T) {
	bo := st()
	x := leg(bo, true)
	if len(x) != 20 {
		t.Fatalf("legal start wanted 20 got %d", len(x))
	}
}

func TestCheckThing(t *testing.T) {
	bo := empty()
	putp(&bo, 7, 4, "K")
	putp(&bo, 0, 4, "k")
	putp(&bo, 5, 4, "r")
	if ischk(bo, true) == false {
		t.Fatalf("rook should check the king")
	}
}

func TestPerftStartDepthTwo(t *testing.T) {
	bo := st()
	if perft(bo, true, 2) != 400 {
		t.Fatalf("bad perft 2 got %d", perft(bo, true, 2))
	}
}

func TestPerftStartDepthThree(t *testing.T) {
	bo := st()
	if perft(bo, true, 3) != 8902 {
		t.Fatalf("bad perft 3 got %d", perft(bo, true, 3))
	}
}

func TestUndoBadMove(t *testing.T) {
	bo := st()
	old := bo
	u := dobad(&bo, m{6, 4, 4, 4})
	undobad(&bo, u)
	if bo != old {
		t.Fatalf("undo did not restore board")
	}
}

func TestSearchGivesMove(t *testing.T) {
	bo := st()
	mv := think(bo, true, 2, 0)
	if mv.a == -1 {
		t.Fatalf("wanted a move")
	}
}

func TestRealBitBoardCount(t *testing.T) {
	bo := st()
	if bitcount(bo) != 32 {
		t.Fatalf("wanted 32 pieces got %d", bitcount(bo))
	}
	if bo.wp == 0 || bo.bk == 0 {
		t.Fatalf("bitboards are not set")
	}
}

func TestFenAndUciMoves(t *testing.T) {
	bo, white := posline("position startpos moves e2e4 e7e5")
	if white != true {
		t.Fatalf("white should be to move")
	}
	if pc(bo, 4, 4) != "P" || pc(bo, 3, 4) != "p" {
		t.Fatalf("uci position did not play moves")
	}
}

func TestCastleMoveOnBitBoard(t *testing.T) {
	bo, white := fen("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	if white == false {
		t.Fatalf("wanted white")
	}
	l := leg(bo, true)
	ok := false
	for i := 0; i < len(l); i++ {
		if same(l[i], m{7, 4, 7, 6}) {
			ok = true
		}
	}
	if ok == false {
		t.Fatalf("wanted castle move")
	}
	domove(&bo, m{7, 4, 7, 6})
	if pc(bo, 7, 6) != "K" || pc(bo, 7, 5) != "R" {
		t.Fatalf("castle did not move king and rook")
	}
}

func TestShortSelfGameDoesNotCrash(t *testing.T) {
	n := selfgame(4)
	if n == 0 {
		t.Fatalf("self game made no moves")
	}
}
