package core

import (
	"errors"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/tree"
)

const queriedName = "database"

type DatabaseCurrent interface {
	GetURL() string
	GetPort() string
	GetReadyQuery() zitadel.EnsureFunc
}

func ParseQueriedForDatabase(queried map[string]interface{}) (DatabaseCurrent, error) {
	queriedDB, ok := queried[queriedName]
	if !ok {
		return nil, errors.New("no current state for database found")
	}
	currentDBTree, ok := queriedDB.(*tree.Tree)
	if !ok {
		return nil, errors.New("current state does not fullfil interface")
	}
	currentDB, ok := currentDBTree.Parsed.(DatabaseCurrent)
	if !ok {
		return nil, errors.New("current state does not fullfil interface")
	}

	return currentDB, nil
}

func SetQueriedForDatabase(queried map[string]interface{}, databaseCurrent *tree.Tree) {
	queried[queriedName] = databaseCurrent
}
