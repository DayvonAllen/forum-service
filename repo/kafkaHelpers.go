package repo

import (
	"example.com/app/domain"
	"fmt"
)

func ProcessMessage(message domain.Message) error {

	if message.ResourceType == "user" {
		// 201 is the created messageType
		if message.MessageType == 201 {
			user := message.User
			err := UserRepoImpl{}.Create(&user)

			if err != nil {
				return err
			}
			return nil
		}

		// 200 is the updated messageType
		if message.MessageType == 200 {
			user := message.User

			err := UserRepoImpl{}.UpdateByID(&user)
			if err != nil {
				return err
			}
			return nil
		}

		// 204 is the deleted messageType
		if message.MessageType == 204 {
			user := message.User

			err := UserRepoImpl{}.DeleteByID(user.Id)

			if err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("cannot process this message")
}



