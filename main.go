package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Team uint8
type Piece uint8

const (
	BOARD_WIDTH  = 9
	BOARD_HEIGHT = 9

	NO_TEAM Team = 0
	BLUE    Team = 'b'
	RED     Team = 'r'
	DRAW    Team = 'd'

	NO_PIECE Piece = 0
	ROCK     Piece = 'R'
	PAPER    Piece = 'P'
	SCISSORS Piece = 'S'

	//blue zone (red's goal) is A1 (bottom left corner)
	//red zone (blue's goal) is I9 (top right corner)
	BLUE_ZONE_X = 0
	BLUE_ZONE_Y = BOARD_HEIGHT - 1
	RED_ZONE_X  = BOARD_WIDTH - 1
	RED_ZONE_Y  = 0
)

var (
	BOTH_TEAMS = [2]Team{BLUE, RED}

	TEAM_NAMES = map[Team]string{
		NO_TEAM: "none",
		BLUE:    "blue",
		RED:     "red",
		DRAW:    "draw",
	}

	PIECE_EMOJI = map[Piece]rune{
		NO_PIECE: ' ',
		ROCK:     '⬠',
		PAPER:    '🗎',
		SCISSORS: '✂',
	}

	ZONE_XY = map[Team][2]int{
		BLUE: {BLUE_ZONE_X, BLUE_ZONE_Y},
		RED:  {RED_ZONE_X, RED_ZONE_Y},
	}

	VALID_CAPTURE = map[Piece]Piece{
		ROCK:     SCISSORS, //rock can only capture scissors
		SCISSORS: PAPER,
		PAPER:    ROCK,
	}

	MOVES_WITHOUT_CAPTURE_FOR_DRAW int64 = 100

	STARTING_PIECES = [][2]string{
		{"bR", "D2"}, //blue rock on D2
		{"bR", "C3"},
		{"bR", "B4"},
		{"bP", "E2"},
		{"bP", "D3"},
		{"bP", "C4"},
		{"bP", "B5"},
		{"bS", "E3"},
		{"bS", "D4"},
		{"bS", "C5"},

		{"rR", "F8"}, //red rock on F8
		{"rR", "G7"},
		{"rR", "H6"},
		{"rP", "E8"},
		{"rP", "F7"},
		{"rP", "G6"},
		{"rP", "H5"},
		{"rS", "E7"},
		{"rS", "F6"},
		{"rS", "G5"},
	}

	FG_BLUE  FGColor = 34
	FG_RED   FGColor = 31
	FG_BLACK FGColor = 30
	FG_WHITE FGColor = 37
	BG_BLUE  BGColor = 44
	BG_RED   BGColor = 41
	BG_BLACK BGColor = 40
	BG_WHITE BGColor = 47
)

type FGColor uint8
type BGColor uint8

func assert(cond bool, v any) {
	if !cond {
		panic(v)
	}
}

func posToXY(pos string) (int, int) {
	assert(len(pos) == 2, pos)
	assert('A' <= pos[0] && pos[0] <= 'I', pos)
	assert('1' <= pos[1] && pos[1] <= '9', pos)
	x := pos[0] - 'A'
	y := 9 - (pos[1] - '0')
	return int(x), int(y)
}

type Square struct {
	team  Team
	piece Piece
}
type Board [BOARD_HEIGHT][BOARD_WIDTH]Square
type Game struct {
	board       Board
	turn        Team
	moveNum     int64
	lastCapture int64
	validMoves  []Move
}

func setupBoard() Board {
	board := Board{}
	for _, v := range STARTING_PIECES {
		team := Team(v[0][0])
		piece := Piece(v[0][1])
		pos := v[1]
		x, y := posToXY(pos)
		board[y][x] = Square{team, piece}
	}
	return board
}

func newGame() *Game {
	game := Game{setupBoard(), BLUE, 0, 0, []Move{}}
	findValidMoves(&game)
	return &game
}

func serializeGame(game *Game) string {
	boardStr := ""
	for _, row := range game.board {
		for _, square := range row {
			if square.team == NO_TEAM {
				boardStr += "_"
			} else {
				boardStr += string(square.team) + string(square.piece)
			}
		}
	}
	return fmt.Sprintf("%d-%d-%s-%s", game.moveNum, game.lastCapture, string(game.turn), boardStr)
}

func parseInt(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}

func deserializeGame(s string) *Game {
	split := strings.Split(s, "-")
	moveNum := parseInt(split[0])
	lastCapture := parseInt(split[1])
	turn := Team(split[2][0])
	boardStr := split[3]
	emptySquare := string([]rune{rune(NO_TEAM), rune(NO_PIECE)})
	boardStr = strings.ReplaceAll(boardStr, "_", emptySquare)
	board := Board{}
	for y, row := range board {
		for x := range row {
			idx := (y*BOARD_WIDTH + x) * 2
			board[y][x] = Square{Team(boardStr[idx]), Piece(boardStr[idx+1])}
		}
	}
	game := Game{board, turn, moveNum, lastCapture, []Move{}}
	findValidMoves(&game)
	return &game
}

func printColor(text string, bg BGColor) {
	fmt.Printf("\x1b[0;%d;%dm%s\x1b[0m", FG_WHITE, bg, text)
}

func printBoard(board Board) {
	for y, row := range board {
		fmt.Print(9-y, " ")
		for x, square := range row {
			team := square.team
			piece := square.piece
			var bg BGColor
			switch team {
			case RED:
				bg = BG_RED
			case BLUE:
				bg = BG_BLUE
			default:
				if (x+y)%2 == 0 {
					bg = BG_BLACK
				} else {
					bg = BG_WHITE
				}
				if x == BLUE_ZONE_X && y == BLUE_ZONE_Y {
					bg = BG_BLUE
				}
				if x == RED_ZONE_X && y == RED_ZONE_Y {
					bg = BG_RED
				}
			}
			printColor(string(PIECE_EMOJI[piece])+" ", bg)
		}
		fmt.Println()
	}
	fmt.Print("  ")
	for x := range 9 {
		fmt.Print(string(rune(x+'A')), " ")
	}
	fmt.Println()
	// fmt.Print("    ")
	// for x := range 9 {
	// 	fmt.Print(x, " ")
	// }
	// fmt.Println()
}

func printEvalBar(evalValue float64, width int, indent int) {
	plus := ""
	if evalValue > 0 {
		plus = "+"
	}
	text := fmt.Sprintf("%s%.2f", plus, evalValue)
	textPos := int((width - len(text)) / 2)
	redBlocks := int((evalValue + 1) / 2 * float64(width))
	fmt.Print(strings.Repeat(" ", indent))
	for i := range width {
		char := " "
		if textPos <= i && i < textPos+len(text) {
			char = string(text[i-textPos])
		}
		bg := BG_BLUE
		if i < redBlocks {
			bg = BG_RED
		}
		printColor(char, bg)
	}
	fmt.Println()
}

func printGame(game *Game, winner Team) {
	fmt.Println(serializeGame(game))
	printBoard(game.board)
	printEvalBar(evaluateGame(game), 18, 2)
	fmt.Printf("move #%d (%d/%d since last capture)\n", game.moveNum, game.moveNum-game.lastCapture, MOVES_WITHOUT_CAPTURE_FOR_DRAW)
	if winner != NO_TEAM {
		var bg BGColor
		switch winner {
		case RED:
			bg = BG_RED
		case BLUE:
			bg = BG_BLUE
		default:
			bg = BG_BLACK
		}
		printColor(fmt.Sprintf("winner is: %s", TEAM_NAMES[winner]), bg)
		fmt.Println()
	} else {
		fmt.Printf("%s to move... ", TEAM_NAMES[game.turn])
	}
}

func oppositeTeam(team Team) Team {
	assert(team == BLUE || team == RED, team)
	if team == BLUE {
		return RED
	} else {
		return BLUE
	}
}

type Move struct {
	startX int
	startY int
	endX   int
	endY   int
}

func findValidMoves(game *Game) {
	validMoves := make([]Move, 0)
	for y := range BOARD_HEIGHT {
		for x := range BOARD_WIDTH {
			square := game.board[y][x]
			team := square.team
			piece := square.piece
			if team == game.turn {
				otherTeam := oppositeTeam(team)
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						if dx != 0 || dy != 0 {
							x2 := x + dx
							y2 := y + dy
							if 0 <= x2 && x2 < BOARD_WIDTH && 0 <= y2 && y2 < BOARD_HEIGHT {
								square2 := game.board[y2][x2]
								team2 := square2.team
								piece2 := square2.piece

								if team2 == 0 || (team2 == otherTeam && piece2 == VALID_CAPTURE[piece]) {
									validMoves = append(validMoves, Move{x, y, x2, y2})
								}
							}
						}
					}
				}
			}
		}
	}
	game.validMoves = validMoves //this should be the only place this is set, would make it immutable after this if I could
	//this way we can copy a game object but keep the reference to the same validMoves for efficiency
	//then when the copied game object updates, we need to recalculate and replace validMoves anyway, so the original is unaffected
}

type Undo struct {
	lastCapture int64
	move        Move
	startSquare Square
	endSquare   Square
}

func makeMove(game *Game, move Move) Undo {
	startX := move.startX
	startY := move.startY
	endX := move.endX
	endY := move.endY
	assert(startX != endX || startY != endY, move)
	undo := Undo{game.lastCapture, move, game.board[startY][startX], game.board[endY][endX]}

	isCapture := game.board[endY][endX].team == oppositeTeam(game.turn)
	game.board[endY][endX] = game.board[startY][startX]
	game.board[startY][startX] = Square{}

	game.moveNum++
	if isCapture {
		game.lastCapture = game.moveNum
	}

	game.turn = oppositeTeam(game.turn)
	findValidMoves(game)
	return undo
}

func undoMove(game *Game, undo Undo) {
	game.board[undo.move.startY][undo.move.startX] = undo.startSquare
	game.board[undo.move.endY][undo.move.endX] = undo.endSquare
	game.moveNum--
	game.lastCapture = undo.lastCapture
	game.turn = oppositeTeam(game.turn)
}

func checkForWin(game *Game) Team {
	if game.board[BLUE_ZONE_Y][BLUE_ZONE_X].team == RED {
		return RED //if red has reached the blue zone, red wins
	}
	if game.board[RED_ZONE_Y][RED_ZONE_X].team == BLUE {
		return BLUE
	}

	//if you have no legal moves, you lose
	if len(game.validMoves) == 0 {
		return oppositeTeam(game.turn)
	}

	//if 100 moves go by with no captures, it's a draw
	if game.moveNum-game.lastCapture >= MOVES_WITHOUT_CAPTURE_FOR_DRAW {
		return DRAW
	}

	return NO_TEAM //if the game is not over, return that no team has won
}

func abs[T int | float64](v T) T {
	if v > 0 {
		return v
	} else {
		return -v
	}
}

func distanceBetween[T int | float64](x1 T, y1 T, x2 T, y2 T) T {
	return max(abs(x2-x1), abs(y2-y1))
}

type EvalData struct {
	pieces   map[Piece]uint8
	total    uint8
	distance float64
	closest  float64
	avgX     float64
	avgY     float64
	cluster  float64
	balance  float64
}

func evaluateGame(game *Game) float64 {
	// -1 means win for blue, +1 means win for red
	// anything in between is a heuristic guess as to which team is closer to winning
	winner := checkForWin(game)
	if winner == BLUE {
		return -1
	}
	if winner == RED {
		return 1
	}
	if winner == DRAW {
		return 0 //draw is always 0 regardless of heuristic
	}

	data := map[Team]*EvalData{}

	for _, team := range BOTH_TEAMS {
		data[team] = &EvalData{}
		data[team].pieces = map[Piece]uint8{}
		data[team].pieces[ROCK] = 0
		data[team].pieces[PAPER] = 0
		data[team].pieces[SCISSORS] = 0
	}

	for y, row := range game.board {
		for x, square := range row {
			team := square.team
			piece := square.piece
			if team != NO_TEAM {
				data[team].pieces[piece] += 1
				data[team].total += 1

				xy := ZONE_XY[oppositeTeam(team)]
				distance := float64(distanceBetween(x, y, xy[0], xy[1]))
				data[team].distance += distance
				data[team].closest = min(data[team].closest, distance)

				data[team].avgX += float64(x)
				data[team].avgY += float64(y)
			}
		}
	}

	for _, team := range BOTH_TEAMS {
		total := float64(data[team].total)
		if total > 0 {
			data[team].distance /= total
			data[team].avgX /= total
			data[team].avgY /= total
		}
	}

	for y, row := range game.board {
		for x, square := range row {
			team := square.team
			if team != NO_TEAM {
				total := float64(data[team].total)
				data[team].cluster += distanceBetween(float64(x), float64(y), data[team].avgX, data[team].avgY) / total
			}
		}
	}

	//the further apart the blue cluster (higher cluster value), the better it is for red
	clusterEval := (data[BLUE].cluster - data[RED].cluster) / ((BOARD_HEIGHT - 1) / 2)

	distanceEval := (data[BLUE].distance - data[RED].distance) / BOARD_HEIGHT
	closestEval := (data[BLUE].closest - data[RED].closest) / BOARD_HEIGHT

	//calculate material advantage
	materialEval := float64(data[RED].total-data[BLUE].total) / float64(len(STARTING_PIECES)/2)

	for _, team := range BOTH_TEAMS {
		pieces := data[team].pieces
		vals := []uint8{pieces[ROCK], pieces[PAPER], pieces[SCISSORS]}
		if slices.Max(vals) > 0 {
			data[team].balance = float64(slices.Min(vals)) / float64(slices.Max(vals))
		}
	}
	matchupEval := data[RED].balance - data[BLUE].balance

	return materialEval*.5 + matchupEval*.1 + distanceEval*.15 + closestEval*.2 + clusterEval*.05
}

func randomChoice[T any](arr []T) T {
	return arr[rand.Intn(len(arr))]
}

func firstMoveEngine(game *Game) Move {
	return game.validMoves[0]
}

func randomEngine(game *Game) Move {
	return randomChoice(game.validMoves)
}

func captureEngine(game *Game) Move {
	for _, move := range game.validMoves {
		isCapture := game.board[move.endY][move.endX].team == oppositeTeam(game.turn)
		if isCapture { //as soon as we find any capture, do it
			return move
		}
	}
	return randomChoice(game.validMoves) //fallback to random moves
}

func capturesEngine(game *Game) Move {
	captures := make([]Move, 0)
	for _, move := range game.validMoves {
		isCapture := game.board[move.endY][move.endX].team == oppositeTeam(game.turn)
		if isCapture {
			captures = append(captures, move)
		}
	}
	if len(captures) > 0 { //if we found any captures, select one at random
		return randomChoice(captures)
	}
	return randomChoice(game.validMoves) //fallback to random moves
}

func minimax(game *Game, myTeam Team, depth uint8) (float64, *Move) {
	isMyTurn := myTeam == game.turn
	score := evaluateGame(game)
	if myTeam == BLUE { //since -1 = win for blue
		score = -score //invert it
	}
	moves := game.validMoves
	if depth == 0 || len(moves) == 0 {
		return score, nil
	}
	bestScore := math.MaxFloat64
	if isMyTurn {
		bestScore = -math.MaxFloat64
	}
	bestMoves := make([]Move, 0)
	for _, move := range moves {
		undo := makeMove(game, move)
		// moves2 := game.validMoves
		if checkForWin(game) == myTeam {
			undoMove(game, undo)
			return 1, &move
		}
		score, _ = minimax(game, myTeam, depth-1)
		if (isMyTurn && score > bestScore) || (!isMyTurn && score < bestScore) {
			bestScore = score
			bestMoves = make([]Move, 0)
			bestMoves = append(bestMoves, move)
		} else if score == bestScore {
			bestMoves = append(bestMoves, move)
		}
		undoMove(game, undo)
	}
	// fmt.Println(bestMoves, len(bestMoves), depth)
	move := randomChoice(bestMoves)
	return bestScore, &move
}
func minimaxEngine(game *Game) Move {
	score, move := minimax(game, game.turn, 3)
	assert(move != nil, score)
	return *move
}

func alphabeta(game *Game, myTeam Team, depth uint8, alpha float64, beta float64) (float64, *Move) {
	isMyTurn := myTeam == game.turn
	score := evaluateGame(game)
	if myTeam == BLUE { //since -1 = win for blue
		score = -score //invert it
	}
	moves := game.validMoves
	if depth == 0 || len(moves) == 0 {
		return score, nil
	}
	bestScore := math.MaxFloat64
	if isMyTurn {
		bestScore = -math.MaxFloat64
	}
	bestMoves := make([]Move, 0)
	for _, move := range moves {
		undo := makeMove(game, move)
		// moves2 := game.validMoves
		if checkForWin(game) == myTeam {
			undoMove(game, undo)
			return 1, &move
		}
		score, _ = alphabeta(game, myTeam, depth-1, alpha, beta)
		if (isMyTurn && score > bestScore) || (!isMyTurn && score < bestScore) {
			bestScore = score
			bestMoves = make([]Move, 0)
			bestMoves = append(bestMoves, move)
		} else if score == bestScore {
			bestMoves = append(bestMoves, move)
		}
		undoMove(game, undo)
		if isMyTurn { //if it is my turn, we are maximizing the value
			if bestScore >= beta {
				break //beta cutoff
			}
			if bestScore > alpha {
				alpha = bestScore
			}
		} else {
			if bestScore <= alpha {
				break //alpha cutoff
			}
			if bestScore < beta {
				beta = bestScore
			}
		}
	}
	// fmt.Println(bestMoves, len(bestMoves), depth)
	move := randomChoice(bestMoves)
	return bestScore, &move
}
func alphabetaEngine(game *Game) Move {
	score, move := alphabeta(game, game.turn, 4, -math.MaxFloat64, math.MaxFloat64)
	assert(move != nil, score)
	return *move
}

func playerEngine(game *Game) Move {
	var input string
	fmt.Print("Your turn! Enter a move in the notation a1a2: ")
	for {
		fmt.Scanf("%v", &input)
		input = strings.ToUpper(input)
		if len(input) == 4 &&
			'A' <= input[0] && input[0] <= 'I' &&
			'1' <= input[1] && input[1] <= '9' &&
			'A' <= input[2] && input[2] <= 'I' &&
			'1' <= input[3] && input[3] <= '9' {
			startX, startY := posToXY(input[0:2])
			endX, endY := posToXY(input[2:4])
			move := Move{startX, startY, endX, endY}
			fmt.Println(move)
			if slices.Contains(game.validMoves, move) {
				return move
			} else {
				fmt.Println("That move is not legal")
			}
		} else {
			fmt.Println("What you just said to me is lowkey indecipherable, stick to the notation bestie")
		}
	}
}

type Engine func(*Game) Move

func runGame(blueEngine Engine, redEngine Engine) {
	game := newGame()
	winner := NO_TEAM
	engines := map[Team]Engine{
		BLUE: blueEngine,
		RED:  redEngine,
	}
	totalTime := map[Team]time.Duration{
		BLUE: 0,
		RED:  0,
	}
	for {
		winner = checkForWin(game)
		if winner != NO_TEAM { //|| game.moveNum == 15 { //TODO tmp
			break
		}
		printGame(game, winner)
		engine := engines[game.turn]
		start := time.Now()
		move := engine(game)
		elapsed := time.Since(start)
		fmt.Printf("%s engine took %v\n", TEAM_NAMES[game.turn], elapsed)
		totalTime[game.turn] += elapsed
		makeMove(game, move)
	}
	printGame(game, winner)

	blueMoves := game.moveNum / 2
	redMoves := (game.moveNum + 1) / 2

	if blueMoves == 0 {
		blueMoves = 1
	}
	if redMoves == 0 {
		redMoves = 1
	}
	fmt.Printf("blue engine total time %v (average %v per move)\n", totalTime[BLUE], totalTime[BLUE]/time.Duration(blueMoves))
	fmt.Printf("red engine total time %v (average %v per move)\n", totalTime[RED], totalTime[RED]/time.Duration(redMoves))
}

func main() {
	fmt.Println("Version", runtime.Version())
	fmt.Println("NumCPU", runtime.NumCPU())
	fmt.Println("GOMAXPROCS", runtime.GOMAXPROCS(0))

	assert(serializeGame(newGame()) == serializeGame(deserializeGame(serializeGame(newGame()))), ' ')
	// printGame(newGame(), NO_TEAM)

	f, err := os.Create("cpu.pb.gz")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

	// runtime.GC() // run GC first to capture only the most current objects
	// f2, err := os.Create("heap.pb.gz")
	// if err != nil {
	// 	panic(err)
	// }
	// defer f2.Close()
	// pprof.Lookup("heap").WriteTo(f2, 0)

	// f3, err := os.Create("allocs.pb.gz")
	// if err != nil {
	// 	panic(err)
	// }
	// defer f3.Close()
	// pprof.Lookup("allocs").WriteTo(f3, 0)

	// runGame(firstMoveEngine, firstMoveEngine)
	// runGame(randomEngine, randomEngine)
	// runGame(randomEngine, firstMoveEngine)
	// runGame(randomEngine, captureEngine)
	// runGame(randomEngine, capturesEngine)
	// runGame(minimaxEngine, captureEngine)
	// runGame(minimaxEngine, minimaxEngine)
	runGame(minimaxEngine, alphabetaEngine)
	// runGame(firstMoveEngine, alphabetaEngine)
	// runGame(playerEngine, alphabetaEngine)
}
