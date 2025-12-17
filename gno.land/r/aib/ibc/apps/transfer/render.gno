package transfer

import (
	"chain"

	"gno.land/p/aib/jsonpage"
	"gno.land/p/nt/mux"
	"gno.land/p/nt/ufmt"
	"gno.land/p/onbloc/json"
)

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", renderHome)
	router.HandleFunc("denoms", renderDenoms)
	router.HandleFunc("denoms/ibc/{hash}", renderDenom)
	router.HandleFunc("total_escrow/{denom}", renderTotalEscrowForDenom)
	// TODO add total escrowed for denom
	return router.Render(path)
}

func renderHome(w *mux.ResponseWriter, r *mux.Request) {
	// TODO list available endpoints ?
	w.Write(`["Hello IBC transfer!"]`)
}

func renderDenoms(w *mux.ResponseWriter, r *mux.Request) {
	renderNode(w, jsonpage.Render(denoms, r, nil))
}

func renderDenom(w *mux.ResponseWriter, r *mux.Request) {
	denom := "ibc/" + r.GetVar("hash")
	d, ok := denoms.Get(denom)
	if !ok {
		renderNode(w, nodeError(ufmt.Sprintf("denom %s not found", denom)))
		return
	}
	renderNode(w, d.(Denom).RenderJSON())
}

func renderTotalEscrowForDenom(w *mux.ResponseWriter, r *mux.Request) {
	denom := r.GetVar("denom")
	var amt int64
	x, found := totalEscrow.Get(denom)
	if found {
		amt = x.(chain.Coin).Amount
	}
	renderNode(w, json.ObjectNode("", map[string]*json.Node{
		"denom":  json.StringNode("", denom),
		"amount": json.NumberNode("", float64(amt)),
	}))
}

func renderNode(w *mux.ResponseWriter, n *json.Node) {
	bz, err := json.Marshal(n)
	if err != nil {
		panic(err)
	}
	w.Write(string(bz))
}

func nodeError(msg string, args ...any) *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"error": json.StringNode("", ufmt.Sprintf(msg, args...)),
	})
}
