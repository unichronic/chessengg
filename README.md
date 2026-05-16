# chessengg

A small chess engine written in Go.

## Run

```sh
go run .
```

Run as a very basic UCI engine:

```sh
go run . uci
```

## Test

```sh
go test ./...
```

## Current Engine Features

- legal move filtering by king safety
- perft
- make/undo move
- minimax and alpha-beta
- rough move ordering
- iterative deepening with a timer
- crude transposition table
- quiescence search
- simple evaluation extras
- bitboard-native board, move generation, legality, perft, search, eval, and UCI position handling
- basic UCI command loop
- short self-play smoke test
