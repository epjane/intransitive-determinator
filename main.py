#!/usr/bin/env python3
import random

BOARD_WIDTH = 9
BOARD_HEIGHT = 9
BLUE_ZONE = 'A1' #bottom left corner
RED_ZONE =  'I9' #top right corner

BLUE = 'b'
RED = 'r'
DRAW = 'd'
COLOR_NAMES = {
    'b': 'blue',
    'r': 'red',
    'd': 'draw',
}

ROCK = 'R'
PAPER = 'P'
SCISSORS = 'S'
EMPTY = '  '

EMOJI = {
    ROCK: '⬠', #🪨⬡
    PAPER: '🗎',
    SCISSORS: '✂',
    ' ': ' ',
}

VALID_CAPTURES = {
    ROCK: [SCISSORS], #rock can only capture scissors
    SCISSORS: [PAPER],
    PAPER: [ROCK],
}

MOVES_WITHOUT_CAPTURE_FOR_DRAW = 100

STARTING_PIECES = [
    ('bR', 'D2'), #blue rock on D2
    ('bR', 'C3'),
    ('bR', 'B4'),
    ('bP', 'E2'),
    ('bP', 'D3'),
    ('bP', 'C4'),
    ('bP', 'B5'),
    ('bS', 'E3'),
    ('bS', 'D4'),
    ('bS', 'C5'),

    ('rR', 'F8'), #red rock on F8
    ('rR', 'G7'),
    ('rR', 'H6'),
    ('rP', 'E8'),
    ('rP', 'F7'),
    ('rP', 'G6'),
    ('rP', 'H5'),
    ('rS', 'E7'),
    ('rS', 'F6'),
    ('rS', 'G5'),
]

ALPHABET = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
assert len(ALPHABET) == 26

def posToXY(pos):
    assert type(pos) == str
    assert len(pos) == 2
    p0, p1 = pos
    assert p0 in ALPHABET
    assert str(int(p1)) == p1
    x = ord(p0)-65 #ord('A')-65 = 0
    y = 9-int(p1)
    assert 0 <= x < BOARD_WIDTH
    assert 0 <= y < BOARD_HEIGHT
    return x, y

BLUE_ZONE_XY = posToXY(BLUE_ZONE)
RED_ZONE_XY = posToXY(RED_ZONE)
assert BLUE_ZONE_XY == (0, 8)
assert RED_ZONE_XY == (8, 0)
ZONE_XY = {
    BLUE: BLUE_ZONE_XY,
    RED: RED_ZONE_XY,
}

def newBoard():
    board = [[EMPTY] * BOARD_WIDTH for _ in range(BOARD_HEIGHT)]
    for piece, pos in STARTING_PIECES:
        x, y = posToXY(pos)
        board[y][x] = piece
    return board

FG_BLUE = 34
FG_RED = 31
FG_BLACK = 30
FG_WHITE = 37
BG_BLUE = 44
BG_RED = 41
BG_BLACK = 40
BG_WHITE = 47
'''
format = ';'.join([str(style), str(fg), str(bg)])
'\x1b[0;31;41m text \x1b[0m'
'''

def printColor(text, bg=BG_BLACK, fg=FG_WHITE):
    print(f'\x1b[0;{fg};{bg}m{text}\x1b[0m', end='')

def printBoard(board):
    for y, row in enumerate(board):
        print(y, 9-y, end=' ')
        for x, (team, piece) in enumerate(row):
            if team == RED:
                bg = BG_RED
            elif team == BLUE:
                bg = BG_BLUE
            else:
                bg = BG_BLACK if (x+y) % 2 == 0 else BG_WHITE
                if x == 0 and y == 8: bg = BG_BLUE
                if x == 8 and y == 0: bg = BG_RED
            printColor(EMOJI[piece]+' ', bg)
        print()
    print(' ', ' ', ' '.join([x for x in ALPHABET[:len(board[0])]]))
    print(' ', ' ', ' '.join([str(x) for x in range(len(board[0]))]))
    # for y, row in enumerate(board):
    #     print(y, 9-y, '|'+'|'.join(row)+'|')
    # print(' ', ' ', ' '.join([' '+x for x in ALPHABET[:len(board[0])]]))
    # print(' ', ' ', ' '.join([' '+str(x) for x in range(len(board[0]))]))

# board = newBoard()
# printBoard(board)

def newGame():
    return {'moveNum': 0, 'lastCapture': 0, 'turn': BLUE, 'board': newBoard()}

def serializeGame(game):
    boardStr = ''.join([''.join(row) for row in game['board']]).replace('  ','_')
    return f"{game['moveNum']}-{game['lastCapture']}-{game['turn']}-{boardStr}"

def deserializeGame(s):
    moveNum, lastCapture, turn, boardStr = s.split('-')
    moveNum = int(moveNum)
    lastCapture = int(lastCapture)
    boardStr = boardStr.replace('_', '  ')
    board = []
    for y in range(BOARD_HEIGHT):
        row = []
        for x in range(BOARD_HEIGHT):
            idx = (y*BOARD_WIDTH+x)*2
            row.append(boardStr[idx:idx+2])
        board.append(row)
    return {'moveNum': moveNum, 'lastCapture': lastCapture, 'turn': turn, 'board': board}

assert newGame() == deserializeGame(serializeGame(newGame()))
# print(serializeGame(newGame()))
# quit()

def printEvalBar(evalValue, width=10, indent=0):
    text = f'{"+" if evalValue >= 0 else ""}{evalValue:0.2f}'
    textPos = int((width - len(text))/2)
    redBlocks = int((evalValue+1)/2*width)
    print(' '*indent, end='')
    for i in range(width):
        if textPos <= i < textPos+len(text):
            char = text[i-textPos]
        else:
            char = ' '
        bg = BG_RED if i < redBlocks else BG_BLUE
        printColor(char, bg)
    print()

    # print(' '*indent, end='')
    # printColor(' ' * redBlocks, BG_RED)
    # printColor(' ' * (width-redBlocks), BG_BLUE)
    # print()

def printGame(game, evalFunc=None, winner=None):
    turn = game['turn']
    moveNum = game['moveNum']
    lastCapture = game['lastCapture']
    assert turn in (BLUE, RED)
    print(serializeGame(game))
    printBoard(game['board'])
    if evalFunc is not None:
        e = evalFunc(game)
        printEvalBar(e, width=18, indent=4)

    print(f'move #{moveNum} ({moveNum - lastCapture}/{MOVES_WITHOUT_CAPTURE_FOR_DRAW} since last capture)')
    if winner is not None:
        if winner == RED:
            bg = BG_RED
        elif winner == BLUE:
            bg = BG_BLUE
        else:
            bg = BG_BLACK
        printColor(f'winner is: {COLOR_NAMES[winner]}', bg)
    else:
        print(f'{COLOR_NAMES[turn]} to move')
    print()

def oppositeTeam(team):
    assert team in (BLUE, RED)
    if team == BLUE:
        return RED
    else:
        return BLUE

def findValidMoves(game):
    turn = game['turn']
    board = game['board']
    validMoves = []
    for y in range(BOARD_HEIGHT):
        for x in range(BOARD_WIDTH):
            team, piece = board[y][x]
            if team == turn:
                otherTeam = oppositeTeam(team)
                for dx in range(-1, 2):
                    for dy in range(-1, 2):
                        if dx != 0 or dy != 0:
                            if 0 <= x+dx < BOARD_WIDTH and 0 <= y+dy < BOARD_HEIGHT:
                                team2, piece2 = board[y+dy][x+dx]
                                if team2 == ' ' or (team2 == otherTeam and piece2 in VALID_CAPTURES[piece]):
                                    # print(x, y, team+piece, x+dx, y+dy)
                                    validMoves.append([(x, y), (x+dx, y+dy)])
    return validMoves

def makeMove(game, move):
    moveNum = game['moveNum']
    lastCapture = game['lastCapture']
    turn = game['turn']
    board = game['board']
    (startX, startY), (endX, endY) = move
    assert (startX, startY) != (endX, endY)

    undo = {
        'squares': (
            (startX, startY, board[startY][startX]),
            (endX, endY, board[endY][endX]),
        ),
        'lastCapture': lastCapture,
    }

    # board = [row[:] for row in board] #duplicate the board to avoid changing the original
    isCapture = board[endY][endX] != EMPTY #check whether this move is a capture
    board[endY][endX] = board[startY][startX] #copy the piece
    board[startY][startX] = EMPTY #make the original space empty

    moveNum += 1 #increment the move counter
    if isCapture:
        lastCapture = moveNum

    turn = oppositeTeam(turn) #it's now the other player's turn
    game['moveNum'] = moveNum
    game['lastCapture'] = lastCapture
    game['turn'] = turn
    # game['board'] = board
    # {'moveNum': moveNum, 'lastCapture': lastCapture, 'turn': turn, 'board': board} #return the new game state
    return game, undo

def undoMove(game, undo):
    for x, y, val in undo['squares']:
        game['board'][y][x] = val
    game['moveNum'] -= 1
    game['lastCapture'] = undo['lastCapture']
    game['turn'] = oppositeTeam(game['turn'])
    return game

def checkForWin(game, moves, draw_moves=MOVES_WITHOUT_CAPTURE_FOR_DRAW):
    #this only checks for simple win/draw conditions
    moveNum = game['moveNum']
    lastCapture = game['lastCapture']
    turn = game['turn']
    board = game['board']

    x, y = BLUE_ZONE_XY
    if board[y][x][0] == RED: #if red has reached the blue zone, red wins
        return RED
    x, y = RED_ZONE_XY
    if board[y][x][0] == BLUE:
        return BLUE

    #if you have no legal moves, you lose
    if len(moves) == 0:
        return oppositeTeam(turn)

    #if 100 moves go by with no captures, it's a draw
    if moveNum - lastCapture >= draw_moves:
        return DRAW

    return None #if the game is not over

#Chebyshev Distance
def distanceBetween(xy1, xy2):
    (x1, y1), (x2, y2) = xy1, xy2
    distance = max(abs(x2-x1), abs(y2-y1))
    # print(x1, y1, x2, y2, ' D:', distance)
    return distance

def combineEval(materialEval, matchupEval, distanceEval, closestEval, clusterEval):
    return materialEval * .7 + matchupEval * .15 + distanceEval * .1 + closestEval * .05

def combineEval2(materialEval, matchupEval, distanceEval, closestEval, clusterEval):
    return materialEval * .5 + matchupEval * .1 + distanceEval * .15 + closestEval * .2 + clusterEval * .05

def evaluateGame(game, combineFunc=combineEval):
    #-1 means win for blue, +1 means win for red
    #anything in between is a heuristic guess as to which team is closer to winning
    board = game['board']
    x, y = BLUE_ZONE_XY
    if board[y][x][0] == RED: #if red has reached the blue zone, red wins
        return 1
    x, y = RED_ZONE_XY
    if board[y][x][0] == BLUE:
        return -1


    pieces = {
        BLUE: {
            ROCK: 0,
            PAPER: 0,
            SCISSORS: 0,
            'total': 0,
            'distance': 0,
            'closest': BOARD_HEIGHT,
            'avgX': 0,
            'avgY': 0,
            'cluster': 0,
        },
        RED: {
            ROCK: 0,
            PAPER: 0,
            SCISSORS: 0,
            'total': 0,
            'distance': 0,
            'closest': BOARD_HEIGHT,
            'avgX': 0,
            'avgY': 0,
            'cluster': 0,
        },
    }

    #count pieces
    for y in range(BOARD_HEIGHT):
        for x in range(BOARD_WIDTH):
            team, piece = board[y][x]
            if team in (BLUE, RED):
                pieces[team][piece] += 1
                pieces[team]['total'] += 1
                distance = distanceBetween((x, y), ZONE_XY[oppositeTeam(team)])
                pieces[team]['distance'] += distance
                pieces[team]['closest'] = min(pieces[team]['closest'], distance)
                pieces[team]['avgX'] += x
                pieces[team]['avgY'] += y
    # print(pieces)

    for team in (BLUE, RED):
        if pieces[team]['total'] > 0:
            pieces[team]['distance'] /= pieces[team]['total']
            pieces[team]['avgX'] /= pieces[team]['total']
            pieces[team]['avgY'] /= pieces[team]['total']

    for y in range(BOARD_HEIGHT):
        for x in range(BOARD_WIDTH):
            team, piece = board[y][x]
            if team in (BLUE, RED):
                pieces[team]['cluster'] += distanceBetween((x, y), (pieces[team]['avgX'], pieces[team]['avgY']))
    
    for team in (BLUE, RED):
        if pieces[team]['total'] > 0:
            pieces[team]['cluster'] /= pieces[team]['total']

    #the further apart the blue cluster (higher cluster value), the better it is for red
    clusterEval = (pieces[BLUE]['cluster'] - pieces[RED]['cluster']) / ((BOARD_HEIGHT-1)/2)

    distanceEval = (pieces[BLUE]['distance'] - pieces[RED]['distance']) / BOARD_HEIGHT
    closestEval = (pieces[BLUE]['closest'] - pieces[RED]['closest']) / BOARD_HEIGHT

    #calculate material advantage
    materialEval = (pieces[RED]['total'] - pieces[BLUE]['total']) / (len(STARTING_PIECES)/2)

    #TODO count macro positional advantage
    #TODO count micro positional advantage
    #TODO account for whose turn it is
    #TODO experiment with different weights between the 3 advantages

    #calculate matchup advantage
    balance = {}
    for team in (BLUE, RED):
        vals = [pieces[team][ROCK], pieces[team][PAPER], pieces[team][SCISSORS]]
        if max(vals) == 0:
            balance[team] = 0
        else:
            balance[team] = min(vals) / max(vals)
    matchupEval = (balance[RED] - balance[BLUE])

    # return closestEval
    # if materialEval == 0:
    #     materialEval = matchupEval * 0.1

    #TODO in some cases matchup eval should be more important
    #e.g. if red has 5 pieces and blue has 4, but red has no scissors - blue might be in the lead

    return combineFunc(materialEval, matchupEval, distanceEval, closestEval, clusterEval)

def firstMoveEngine(game, moves):
    return moves[0]

def randomEngine(game, moves):
    return random.choice(moves)

def captureEngine(game, moves):
    board = game['board']
    for move in moves:
        (startX, startY), (endX, endY) = move
        isCapture = board[endY][endX] != EMPTY #check whether this move is a capture
        if isCapture:
            return move
    return random.choice(moves) #fallback to random moves

# def runnerEngine(game, moves):
#     #TODO run toward enemy's zone
#     return random.choice(moves) #fallback to random moves

def minimaxEngine(game, moves):
    def minimax(game, moves, myTeam, depth):
        isMyTurn = myTeam == game['turn']
        score = evaluateGame(game)
        if myTeam == BLUE: #since -1 = win for blue
            score = -score #invert it
        if depth == 0 or len(moves) == 0: # or abs(score) > 0.4 #causes crash cuz None
            return score, None
        else:
            bestScore = None
            bestMoves = []
            for move in moves:
                game, undo = makeMove(game, move)
                moves2 = findValidMoves(game)
                if checkForWin(game, moves2) == myTeam:
                    undoMove(game, undo)
                    return 1, move
                score, _ = minimax(game, moves2, myTeam, depth-1)
                # if depth > 2: print(' '*(3-depth), move, score)
                if bestScore is None or (isMyTurn and score > bestScore) or (not isMyTurn and score < bestScore):
                    bestScore = score
                    bestMoves = [move]
                elif score == bestScore:
                    bestMoves.append(move)
                undoMove(game, undo)
            # if depth > 2: print(bestMoves)
            return bestScore, random.choice(bestMoves)

    score, move = minimax(game, moves, game['turn'], depth=2)
    assert move is not None, score
    # print(f'{score=}')
    return move


def alphabetaEngine(game, moves):
    def alphabeta(game, moves, myTeam, depth, alpha=-1000, beta=1000):
        isMyTurn = myTeam == game['turn']
        score = evaluateGame(game, combineFunc=combineEval)
        if myTeam == BLUE: #since -1 = win for blue
            score = -score #invert it
        if depth == 0 or len(moves) == 0:
            return score, None
        else:
            bestScore = None
            bestMoves = []
            for move in moves:
                game, undo = makeMove(game, move)
                moves2 = findValidMoves(game)
                winner = checkForWin(game, moves2, draw_moves=25)
                if winner is not None:
                    undoMove(game, undo)
                    if winner == myTeam: return 1, move
                    elif winner == oppositeTeam(myTeam): return -1, move
                    elif winner == DRAW: return 0, move
                score, _ = alphabeta(game, moves2, myTeam, depth-1, alpha, beta)
                score *= 0.999 #decay at each depth to slightly prefer shorter paths
                # if depth > 2: print(' '*(3-depth), move, score)
                if bestScore is None or (isMyTurn and score > bestScore) or (not isMyTurn and score < bestScore):
                    bestScore = score
                    bestMoves = [move]
                elif score == bestScore:
                    bestMoves.append(move)
                undoMove(game, undo)
                if isMyTurn: #if it is my turn, we are maximizing the value
                    if bestScore >= beta: break #beta cutoff
                    if bestScore > alpha: alpha = bestScore
                else:
                    if bestScore <= alpha: break #alpha cutoff
                    if bestScore < beta: beta = bestScore
            # if depth > 2: print(bestMoves)
            return bestScore, random.choice(bestMoves)

    score, move = alphabeta(game, moves, game['turn'], depth=3)
    assert move is not None, score
    # print(f'{score=}')
    return move

def runGame(blueEngine, redEngine):
    game = newGame()
    # game = deserializeGame('28-25-b-______________rRrR___bP_bS_________rS_rP___bR___rSrSrRrP__bPbP_bS_______bS_______bRbP______bR______')
    # game = deserializeGame('99-54-r-_____rP___bP____rR_rP__rP_bS__rR_rP__bP__rSrS_rR___bS_bR_rS_bR_____bS____________bR________________')
    # game = deserializeGame('45-42-r-________________________rR_______bP______bPrS_rP____bPbSbS______bRbR_______________________')
    winner = None
    while True:
        moves = findValidMoves(game)
        winner = checkForWin(game, moves)
        if winner is not None:
            break
        printGame(game, evaluateGame)
        # print(game['moveNum'], game['lastCapture'])
        engine = blueEngine if game['turn'] == BLUE else redEngine
        game, _ = makeMove(game, engine(game, moves))
        # input()
    printGame(game, None, winner)

# printGame(newGame())
# quit()

# runGame(randomEngine, randomEngine)
# runGame(randomEngine, firstMoveEngine)
# runGame(randomEngine, captureEngine)
# runGame(minimaxEngine, captureEngine)
# runGame(minimaxEngine, minimaxEngine)
runGame(minimaxEngine, alphabetaEngine)

#TODO unit test for 100 move draw rule
