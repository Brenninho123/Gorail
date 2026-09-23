package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Item struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type Project struct {
	Name    string
	Port    string
	items   map[int]Item
	nextID  int
	mu      sync.Mutex
	router  *http.ServeMux
}

func NewProject(name, port string) *Project {
	p := &Project{
		Name:   name,
		Port:   port,
		items:  make(map[int]Item),
		nextID: 1,
		router: http.NewServeMux(),
	}
	p.routes()
	return p
}

func (p *Project) routes() {
	p.router.HandleFunc("GET /items", p.handleListItems)
	p.router.HandleFunc("GET /items/{id}", p.handleGetItem)
	p.router.HandleFunc("POST /items", p.handleCreateItem)
	p.router.HandleFunc("DELETE /items/{id}", p.handleDeleteItem)
}

func (p *Project) Run() error {
	fmt.Printf("Gorail engine running on %s\n", p.Port)
	return http.ListenAndServe(p.Port, p.router)
}

func (p *Project) handleListItems(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	defer p.mu.Unlock()

	list := make([]Item, 0, len(p.items))
	for _, item := range p.items {
		list = append(list, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (p *Project) handleGetItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	p.mu.Lock()
	item, exists := p.items[id]
	p.mu.Unlock()

	if !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (p *Project) handleCreateItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	p.mu.Lock()
	item.ID = p.nextID
	p.nextID++
	p.items[item.ID] = item
	p.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (p *Project) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	p.mu.Lock()
	if _, exists := p.items[id]; !exists {
		p.mu.Unlock()
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	delete(p.items, id)
	p.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	app := NewProject("Gorail", ":8080")
	if err := app.Run(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
