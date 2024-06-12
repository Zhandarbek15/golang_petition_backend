package websocket

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	"petition_api/internal/app/models"
	repository "petition_api/internal/app/repositories"
	"strconv"
	"sync"
)

type VoteWebsocket struct {
	voteRepo repository.VoteRepository
	logger   *logrus.Logger
}

type Message struct {
	MessageType string      `json:"messageType"`
	Payload     interface{} `json:"payload"`
}

func NewVoteWebsocket(voteRepo repository.VoteRepository, logger *logrus.Logger) *VoteWebsocket {
	return &VoteWebsocket{
		voteRepo: voteRepo,
		logger:   logger,
	}
}

func (vw *VoteWebsocket) AddToRoute(route *gin.RouterGroup) {
	route.GET("/ws/:petitionID", vw.handleWebSocket)
}

var (
	upgrade = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	mutex   = &sync.Mutex{}
	clients = make(map[uint]map[*websocket.Conn]bool)
)

func (vw *VoteWebsocket) handleWebSocket(c *gin.Context) {
	petitionIDStr := c.Param("petitionID")
	petitionID, err := strconv.ParseUint(petitionIDStr, 10, 32)
	if err != nil {
		vw.logger.Errorf("Invalid petition ID: %v", err)
		return
	}

	conn, err := upgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		vw.logger.Errorf("Failed to set websocket upgrade: %v", err)
		return
	}

	mutex.Lock()
	if clients[uint(petitionID)] == nil {
		clients[uint(petitionID)] = make(map[*websocket.Conn]bool)
	}
	clients[uint(petitionID)][conn] = true
	mutex.Unlock()

	defer func() {
		mutex.Lock()
		delete(clients[uint(petitionID)], conn)
		mutex.Unlock()
		if err := conn.Close(); err != nil {
			vw.logger.Errorf("Failed to close websocket connection: %v", err)
		}
	}()

	count, err := vw.voteRepo.GetCountVoteByPetitionID(uint(petitionID))
	if err != nil {
		vw.logger.Errorf("Failed to get broadcast vote count: %v", err)
		return
	}

	_ = conn.WriteJSON(Message{
		MessageType: "vote_count",
		Payload:     map[string]int64{"vote_count": count},
	})

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			vw.logger.Errorf("Failed to read message: %v", err)
			_ = conn.WriteJSON(Message{
				MessageType: "error",
				Payload:     map[string]string{"errorMsg": "Failed to read message"},
			})
			continue
		}

		vw.logger.Debugf("Received message: %v", msg.Payload)
		var vote models.Vote
		data, _ := json.Marshal(msg.Payload)
		if err := json.Unmarshal(data, &vote); err != nil {
			vw.logger.Errorf("Failed to unmarshal vote: %v", err)
			_ = conn.WriteJSON(Message{
				MessageType: "error",
				Payload:     map[string]string{"errorMsg": "Invalid payload format"},
			})
			continue
		}

		switch msg.MessageType {
		case "vote":
			if err := vw.voteHandle(&vote); err != nil {
				_ = conn.WriteJSON(Message{
					MessageType: "error",
					Payload:     map[string]string{"errorMsg": err.Error()},
				})
			}
		case "unvote":
			if err := vw.unvoteHandle(&vote); err != nil {
				_ = conn.WriteJSON(Message{
					MessageType: "error",
					Payload:     map[string]string{"errorMsg": err.Error()},
				})
			}
		case "checkVote":
			if err := vw.checkVoteHandle(conn, &vote); err != nil {
				_ = conn.WriteJSON(Message{
					MessageType: "error",
					Payload:     map[string]string{"errorMsg": err.Error()},
				})
			}
		case "closeConn":
			mutex.Lock()
			delete(clients[uint(petitionID)], conn)
			mutex.Unlock()
			if err := conn.Close(); err != nil {
				vw.logger.Errorf("Failed to close websocket connection: %v", err)
			}
			vw.logger.Debug("Closing connection")
			return
		default:
			_ = conn.WriteJSON(Message{
				MessageType: "error",
				Payload:     map[string]string{"errorMsg": "Unknown message type"},
			})
		}
	}
}

func (vw *VoteWebsocket) voteValidate(vote *models.Vote) error {
	if vote.PetitionID == 0 {
		return errors.New("invalid petition ID")
	}
	if vote.UserID == 0 {
		return errors.New("invalid user ID")
	}
	return nil
}

func (vw *VoteWebsocket) voteHandle(vote *models.Vote) error {
	if err := vw.voteValidate(vote); err != nil {
		return err
	}

	_, err := vw.voteRepo.Create(vote)
	if err != nil {
		vw.logger.Errorf("Failed to create vote: %v", err)
		return err
	}

	if err := vw.broadcastVoteCount(vote.PetitionID); err != nil {
		vw.logger.Errorf("Failed to broadcast vote count: %v", err)
		return err
	}

	return nil
}

func (vw *VoteWebsocket) unvoteHandle(vote *models.Vote) error {
	if err := vw.voteValidate(vote); err != nil {
		return err
	}

	err := vw.voteRepo.DeleteByUserIDAndPetitionID(vote.UserID, vote.PetitionID)
	if err != nil {
		vw.logger.Errorf("Failed to delete vote: %v", err)
		return err
	}

	if err := vw.broadcastVoteCount(vote.PetitionID); err != nil {
		vw.logger.Errorf("Failed to broadcast vote count: %v", err)
		return err
	}

	return nil
}

func (vw *VoteWebsocket) checkVoteHandle(conn *websocket.Conn, vote *models.Vote) error {
	if err := vw.voteValidate(vote); err != nil {
		return err
	}

	exist, err := vw.voteRepo.VoteExist(vote.PetitionID, vote.UserID)
	if err != nil {
		vw.logger.Errorf("Failed to check vote: %v", err)
		return err
	}

	if err := conn.WriteJSON(Message{
		MessageType: "checkVote",
		Payload:     map[string]bool{"exists": exist},
	}); err != nil {
		vw.logger.Errorf("Failed to write success message: %v", err)
		return err
	}

	return nil
}

// broadcastVoteCount Чтобы отправить и другим пользовотельям
func (vw *VoteWebsocket) broadcastVoteCount(petitionID uint) error {
	count, err := vw.voteRepo.GetCountVoteByPetitionID(petitionID)
	if err != nil {
		vw.logger.Errorf("Failed to get broadcast vote count: %v", err)
		return err
	}

	mutex.Lock()
	defer mutex.Unlock()
	for client := range clients[petitionID] {
		if err := client.WriteJSON(Message{
			MessageType: "vote_count",
			Payload:     map[string]int64{"vote_count": count},
		}); err != nil {
			vw.logger.Errorf("Write error: %v", err)
			if err := client.Close(); err != nil {
				vw.logger.Errorf("Failed to close client connection: %v", err)
			}
			delete(clients[petitionID], client)
		}
	}

	return nil
}
