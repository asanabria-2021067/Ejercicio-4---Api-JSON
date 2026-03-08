package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"go-http/models"
)

const dataFile = "data/winners.json"

type Store struct {
	mu      sync.RWMutex
	winners []models.Winner
	nextID  int
}

func New() (*Store, error) {
	s := &Store{}
	if err := s.load(); err != nil {
		return nil, fmt.Errorf("loading store: %w", err)
	}
	for _, w := range s.winners {
		if w.ID >= s.nextID {
			s.nextID = w.ID + 1
		}
	}
	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(dataFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.winners)
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.winners, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataFile, data, 0644)
}

func (s *Store) GetAll() []models.Winner {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Winner, len(s.winners))
	copy(out, s.winners)
	return out
}

func (s *Store) GetByID(id int) (models.Winner, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, w := range s.winners {
		if w.ID == id {
			return w, true
		}
	}
	return models.Winner{}, false
}

func (s *Store) Add(w models.Winner) (models.Winner, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w.ID = s.nextID
	s.nextID++
	s.winners = append(s.winners, w)
	return w, s.save()
}

func (s *Store) Replace(id int, w models.Winner) (models.Winner, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, existing := range s.winners {
		if existing.ID == id {
			w.ID = id
			s.winners[i] = w
			return w, true, s.save()
		}
	}
	return models.Winner{}, false, nil
}

func (s *Store) Patch(id int, fields map[string]interface{}) (models.Winner, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, w := range s.winners {
		if w.ID == id {
			if v, ok := fields["player"]; ok {
				str, ok := v.(string)
				if !ok {
					return models.Winner{}, false, errors.New("player must be a string")
				}
				w.Player = str
			}
			if v, ok := fields["nationality"]; ok {
				str, ok := v.(string)
				if !ok {
					return models.Winner{}, false, errors.New("nationality must be a string")
				}
				w.Nationality = str
			}
			if v, ok := fields["club"]; ok {
				str, ok := v.(string)
				if !ok {
					return models.Winner{}, false, errors.New("club must be a string")
				}
				w.Club = str
			}
			if v, ok := fields["year"]; ok {
				f, ok := v.(float64)
				if !ok {
					return models.Winner{}, false, errors.New("year must be a number")
				}
				w.Year = int(f)
			}
			if v, ok := fields["votes"]; ok {
				f, ok := v.(float64)
				if !ok {
					return models.Winner{}, false, errors.New("votes must be a number")
				}
				w.Votes = int(f)
			}
			if v, ok := fields["position"]; ok {
				str, ok := v.(string)
				if !ok {
					return models.Winner{}, false, errors.New("position must be a string")
				}
				w.Position = str
			}
			if v, ok := fields["goals_that_season"]; ok {
				f, ok := v.(float64)
				if !ok {
					return models.Winner{}, false, errors.New("goals_that_season must be a number")
				}
				w.GoalsThatSeason = int(f)
			}
			s.winners[i] = w
			return w, true, s.save()
		}
	}
	return models.Winner{}, false, nil
}

func (s *Store) Delete(id int) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, w := range s.winners {
		if w.ID == id {
			s.winners = append(s.winners[:i], s.winners[i+1:]...)
			return true, s.save()
		}
	}
	return false, nil
}
