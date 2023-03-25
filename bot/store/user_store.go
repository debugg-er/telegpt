package store

import (
	"context"

	"telegpt/bot/config"
	"telegpt/bot/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

type (
	UserStore struct {
		collection *firestore.CollectionRef
	}
)

func NewUserStore() (*UserStore, error) {
	conf := config.Get()
	opt := option.WithCredentialsJSON([]byte(conf.FirebaseCredential))
	client, err := firestore.NewClient(context.Background(), conf.FirebaseProjectId, opt)
	if err != nil {
		return nil, err
	}
	return &UserStore{collection: client.Collection(conf.FirebaseTokenCollectionName)}, nil
}

func (t UserStore) GetUserById(userId string) (*models.User, error) {
	doc, err := t.getUserDocById(userId)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (t UserStore) SetUser(user models.User) error {
	userDoc, err := t.getUserDocById(user.UserId)
	if err != nil {
		return err
	}
	if userDoc != nil {
		updation := []firestore.Update{
			{Path: "token", Value: user.UserOpenAIToken},
		}
		if user.GptMessages != nil {
			updation = append(updation, firestore.Update{
				Path: "gptMessageHistory", Value: user.GptMessages,
			})
		}
		if _, err := userDoc.Ref.Update(context.Background(), updation); err != nil {
			return err
		}
	} else {
		u := t.collection.Doc(user.UserId)
		_, err := u.Create(context.Background(), user)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t UserStore) getUserDocById(userId string) (*firestore.DocumentSnapshot, error) {
	docs := t.collection.Where("userId", "==", userId).Documents(context.Background())
	tokenEntries, err := docs.GetAll()
	if err != nil {
		return nil, err
	}
	if len(tokenEntries) == 0 {
		return nil, nil
	}
	return tokenEntries[0], nil
}
