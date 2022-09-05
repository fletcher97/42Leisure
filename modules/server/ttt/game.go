package ttt

import (
	"42Leisure/server/coms"
	"42Leisure/server/db"
	"42Leisure/server/models"
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Game models.Game_ttt
type Board models.Board

///////////////////////////////////// Utils ////////////////////////////////////

// Save game to DB.
func (g Game) saveGame() error {
	return db.Update(g)
}

// Convert a Game object into a byte slice to be sent to the client.
func (g Game) serializeGame() []byte {
	byteBuffer := new(bytes.Buffer)
	json.NewEncoder(byteBuffer).Encode(g)
	return byteBuffer.Bytes()
}

// Return an empty board.
func getEmptyBoard() models.Board {
	return models.Board{
		models.GameOptionEmpty, models.GameOptionEmpty, models.GameOptionEmpty,
		models.GameOptionEmpty, models.GameOptionEmpty, models.GameOptionEmpty,
		models.GameOptionEmpty, models.GameOptionEmpty, models.GameOptionEmpty}
}

// Create a new Game object with default params ready for to start a game.
func (g *Game) newGame() Game {
	g.First = models.First
	g.Second = models.Second
	g.NextP = models.First
	g.BigBoard = getEmptyBoard()
	for i := 0; i < len(g.Boards); i++ {
		g.Boards[i] = getEmptyBoard()
	}
	g.State = models.GameOptionStart
	g.PlayBoard = 0
	return *g
}

///////////////////////////////////// Game /////////////////////////////////////

// Check if a board for a winner or draw. If one of these states is true return
// a GameOption representing the state.
func (b Board) checkBoard() (state models.GameOption) {
	state = models.GameOptionEmpty
	/* Horizontal check */
	if (b[0] == models.GameOptionX ||
		b[0] == models.GameOptionO) &&
		b[0] == b[1] && b[1] == b[2] {
		state = b[0]
	} else if (b[3] == models.GameOptionX ||
		b[3] == models.GameOptionO) &&
		b[3] == b[4] && b[4] == b[5] {
		state = b[3]
	} else if (b[6] == models.GameOptionX ||
		b[6] == models.GameOptionO) &&
		b[6] == b[7] && b[7] == b[8] {
		state = b[6]
	} else

	/* Vertical check */
	if (b[0] == models.GameOptionX ||
		b[0] == models.GameOptionO) &&
		b[0] == b[3] && b[3] == b[6] {
		state = b[0]
	} else if (b[1] == models.GameOptionX ||
		b[1] == models.GameOptionO) &&
		b[1] == b[4] && b[4] == b[7] {
		state = b[1]
	} else if (b[2] == models.GameOptionX ||
		b[2] == models.GameOptionO) &&
		b[2] == b[5] && b[5] == b[8] {
		state = b[2]
	} else

	/* Diagonal check */
	if (b[0] == models.GameOptionX ||
		b[0] == models.GameOptionO) &&
		b[0] == b[4] && b[4] == b[8] {
		state = b[0]
	} else if (b[2] == models.GameOptionX ||
		b[2] == models.GameOptionO) &&
		b[2] == b[4] && b[4] == b[6] {
		state = b[2]
	} else

	/* Draw check */
	if b[0] != models.GameOptionEmpty &&
		b[1] != models.GameOptionEmpty &&
		b[2] != models.GameOptionEmpty &&
		b[3] != models.GameOptionEmpty &&
		b[4] != models.GameOptionEmpty &&
		b[5] != models.GameOptionEmpty &&
		b[6] != models.GameOptionEmpty &&
		b[7] != models.GameOptionEmpty &&
		b[8] != models.GameOptionEmpty {
		state = models.GameOptionDraw
	}

	return
}

// Check game for winners/draws on small and big boards
func (g *Game) updateGame() {
	var change bool = false
	for i, v := range g.BigBoard {
		if v == models.GameOptionEmpty {
			g.BigBoard[i] = Board(g.Boards[i]).checkBoard()
			change = true
		}
	}
	if change {
		g.State = Board(g.BigBoard).checkBoard()
	}
}

// Try to make a move on the board.
func (g *Game) play(p *models.Player, move uint8) (played bool) {
	played = false
	// First move
	if g.State == models.GameOptionStart && p.Equals(g.P1) {
		g.State = models.GameOptionEmpty
		g.PlayBoard = move
		played = true
	} else

	// Game blocked
	if (g.State == models.GameOptionBlocked &&
		p.Equals(g.P1) && g.NextP == models.Second) ||
		(g.State == models.GameOptionBlocked &&
			p.Equals(g.P2) && g.NextP == models.First) {
		g.State = models.GameOptionEmpty
		g.PlayBoard = move
		played = true
	} else

	// Normal play
	if g.State == models.GameOptionEmpty &&
		((g.NextP == g.First && p.Equals(g.P1)) ||
			(g.NextP == g.Second && p.Equals(g.P2))) {
		if g.Boards[g.PlayBoard][move] == models.GameOptionEmpty {
			if p.Equals(g.P1) {
				g.Boards[g.PlayBoard][move] = g.First
			} else {
				g.Boards[g.PlayBoard][move] = g.Second
			}
			g.updateGame()
			if g.State == models.GameOptionEmpty &&
				g.BigBoard[move] != models.GameOptionEmpty {
				g.State = models.GameOptionBlocked
			}
			played = true
		}
	}

	if played {
		g.saveGame()
	}
	return
}

//////////////////////////////////// Handler ///////////////////////////////////

// List of user currently connected to the server
var users []string

// Crude login method. Must be reworked
func login(user string) *models.Player {

	var p *models.Player
	p.Name = user
	err := db.Db().FirstOrCreate(p).Error
	if err != nil {
		log.Println("login:", err)
		return nil
	}
	users = append(users, user)
	return p
}

// Logout method
func logout(user string) {
	for i := 0; i < len(users); i++ {
		if users[i] == user {
			users[i] = users[len(users)-1]
			users = users[:len(users)-1]
			break
		}
	}
}

type GameInstance struct {
	g      Game
	convP1 chan []byte
	convP2 chan []byte
}

var games []GameInstance = nil

// Create a new game instance
func createNewGame(player *models.Player) *GameInstance {
	newG := (&Game{P1: *player}).newGame()
	newG.saveGame()
	games = append(games,
		GameInstance{newG, make(chan []byte), make(chan []byte)})
	return &games[len(games)-1]
}

// Fetch game instance based on players
func getGame(playerName string) *GameInstance {
	for i := 0; i < len(games); i++ {
		if games[i].g.P1.Name == playerName ||
			games[i].g.P2.Name == playerName {
			return &games[i]
		}
	}
	return nil
}

// Join ongoing game
func joinGame(player *models.Player, playerName string) *GameInstance {
	var g *GameInstance
	if len(playerName) == 0 {
		return getGame(player.Name)
	}
	g = getGame(string(playerName))
	if g == nil {
		return nil
	}
	if (!isPlaying(*player) && g.g.P2.Equals(models.Player{})) {
		g.g.P2 = *player
		g.g.saveGame()
		return g
	}
	if isPlaying(*player) &&
		(g.g.P1.Equals(*player) ||
			g.g.P2.Equals(*player)) {
		return g
	} else {
		return nil
	}
}

func isPlaying(p models.Player) bool {
	for _, g := range games {
		if g.g.P1.Equals(p) || g.g.P2.Equals(p) {
			return true
		}
	}
	return false
}

func TTT(w http.ResponseWriter, r *http.Request) {
	c, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	_, message, _ := c.ReadMessage()
	user := string(message)

	var player *models.Player
	for player == nil {
		player = login(user)
	}
	defer logout(user)

	msg := make(chan []byte)
	go func() { // Receive msg from client
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			msg <- message
		}
	}()

	var game *GameInstance
	var sendCh chan []byte
	var recvCh chan []byte

	defer c.Close()
	for {
		select {
		case message := <-msg:
			switch models.MsgType(message[0]) {
			case models.Ttt_join:
				if game != nil {
					coms.Send(c, models.Ttt_failed.Bytes())
					continue
				}
				game = joinGame(player, string(message[1:]))
				if game == nil {
					coms.Send(c, models.Ttt_failed.Bytes())
				} else {
					if game.g.P1.Equals(*player) {
						sendCh = game.convP1
						recvCh = game.convP2
					} else {
						sendCh = game.convP2
						recvCh = game.convP1
					}
					coms.Send(c, game.g.serializeGame())
				}
			case models.Ttt_create:
				if game != nil {
					coms.Send(c, models.Ttt_failed.Bytes())
					continue
				}
				if !isPlaying(*player) {
					game = createNewGame(player)
					coms.Send(c, game.g.serializeGame())
				} else {
					coms.Send(c, models.Ttt_failed.Bytes())
				}
			case models.Ttt_list:
				for _, g := range games {
					if g.g.P2.Equals(models.Player{}) {
						coms.Send(c, append(
							models.Ttt_list.Bytes(),
							[]byte(g.g.P1.Name)...))
					}
				}
				coms.Send(c, models.Ttt_ok.Bytes())
			case models.Ttt_quit:
				if game != nil {
					sendCh <- models.Ttt_quit.Bytes()
				}
				return
			case models.Ttt_giveUp:
				if game == nil {
					coms.Send(c, models.Ttt_failed.Bytes())
					continue
				}
				if game.g.P1.Equals(*player) {
					game.g.State = game.g.Second
				} else {
					game.g.State = game.g.First
				}
				game.g.saveGame()
				sendCh <- models.Ttt_giveUp.Bytes()
				unloadGames()
			case models.Ttt_play:
				if game == nil {
					coms.Send(c, models.Ttt_failed.Bytes())
					continue
				}
				if !game.g.play(player, message[1]) {
					coms.Send(c, models.Ttt_badMove.Bytes())
				} else {
					coms.Send(c, append(models.Ttt_ok.Bytes(), game.g.serializeGame()...))
				}
			default:
				coms.Send(c, []byte("Bad request"))
			}
		case msg := <-recvCh:
			switch models.MsgType(msg[0]) {
			case models.Ttt_giveUp:
				coms.Send(c, models.Ttt_giveUp.Bytes())
				game = nil
			case models.Ttt_quit:
				coms.Send(c, models.Ttt_quit.Bytes())
			case models.Ttt_play:
				coms.Send(c, game.g.serializeGame())
			}
		}
	}

}

func unloadGames() {
UNLOAD_GAMES_BEGINNING:
	for i, g := range games {
		if g.g.State != models.GameOptionStart &&
			g.g.State != models.GameOptionBlocked &&
			g.g.State != models.GameOptionEmpty {
			games[i] = games[len(games)-1]
			games = games[:len(games)-1]
			goto UNLOAD_GAMES_BEGINNING
		}
	}
}

func LoadGames() {
	var gs []models.Game_ttt

	err := db.Db().Where("state = (?) OR state = (?) OR state = (?)",
		models.GameOptionEmpty,
		models.GameOptionBlocked,
		models.GameOptionStart).Find(&gs).Error

	if err != nil {
		log.Println("loadGames:", err)
	}
	for _, g := range gs {
		games = append(games,
			GameInstance{Game(g), make(chan []byte), make(chan []byte)})
	}
}
