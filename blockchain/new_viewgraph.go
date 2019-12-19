package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/chain"
	"os"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
)

type ViewNode struct {
	view chain.ChainViewInterface
	next map[common.Hash]*ViewNode
	prev *ViewNode
}
type ViewGraph struct {
	name        string
	root        *ViewNode
	node        map[common.Hash]*ViewNode
	leaf        map[common.Hash]*ViewNode
	edgeStr     string
	bestView    *ViewNode
	confirmView *ViewNode
	lock        *sync.RWMutex
}

func NewViewGraph(name string, rootView chain.ChainViewInterface, lock *sync.RWMutex) *ViewGraph {
	s := &ViewGraph{name: name, lock: lock}
	s.leaf = make(map[common.Hash]*ViewNode)
	s.node = make(map[common.Hash]*ViewNode)
	s.root = &ViewNode{
		view: rootView,
		next: make(map[common.Hash]*ViewNode),
		prev: nil,
	}
	s.leaf[*rootView.GetTipBlock().Hash()] = s.root
	s.node[*rootView.GetTipBlock().Hash()] = s.root
	s.confirmView = s.root
	return s
}

func (s *ViewGraph) AddView(b chain.ChainViewInterface) {
	newBlockHash := *b.GetTipBlock().Hash()
	for h, v := range s.node {
		if h == *b.GetTipBlock().GetPreviousViewHash() {
			delete(s.leaf, h)
			s.leaf[newBlockHash] = &ViewNode{
				view: b,
				next: make(map[common.Hash]*ViewNode),
				prev: v,
			}
			v.next[newBlockHash] = s.leaf[newBlockHash]
			s.node[newBlockHash] = s.leaf[newBlockHash]
		}
	}
	s.update()
}

func (s *ViewGraph) update() {
	s.traverse(s.root)
	s.updateConfirmBlock(s.bestView)
}

func (s *ViewGraph) GetBestView() chain.ChainViewInterface {
	return s.bestView.view
}

func (s *ViewGraph) GetFinalView() chain.ChainViewInterface {
	return s.confirmView.view
}

func (s *ViewGraph) Print() {
	s.edgeStr = ""
	s.traverse(s.root)

	dotContent := `digraph {
node [shape=record];
//    rankdir="LR";
newrank=true;
`
	maxTimeSlot := uint64(0)
	for k, v := range s.node {
		shortK := k.String()[0:5]
		dotContent += fmt.Sprintf(`%s_%d_%s [label = "%d:%s"]`, s.name, v.view.GetTipBlock().GetHeight(), string(shortK), v.view.GetTipBlock().GetHeight(), string(shortK)) + "\n"
		dotContent += fmt.Sprintf(`{rank=same; %s_%d_%s; slot_%d;}`, s.name, v.view.GetTipBlock().GetHeight(), string(shortK), v.view.GetTipBlock()) + "\n"
		if v.view.GetTimeslot() > maxTimeSlot {
			maxTimeSlot = v.view.GetTimeslot()
		}
	}

	for i := uint64(0); i < maxTimeSlot; i++ {
		dotContent += fmt.Sprintf("slot_%d -> slot_%d;", i, i+1) + "\n"
	}

	dotContent += s.edgeStr
	dotContent += `}`

	fd, _ := os.OpenFile(s.name+".dot", os.O_WRONLY|os.O_CREATE, 0666)
	fd.Truncate(0)
	fd.Write([]byte(dotContent))
	fd.Close()

}

func (s *ViewGraph) traverse(n *ViewNode) {
	if n.next != nil && len(n.next) != 0 {
		for h, v := range n.next {
			s.edgeStr += fmt.Sprintf("%s_%d_%s -> %s_%d_%s;\n", s.name, n.view.GetTipBlock().GetHeight(), string(n.view.GetTipBlock().Hash().String()[0:5]), s.name, v.view.GetTipBlock().GetHeight(), string(h.String()[0:5]))
			s.traverse(v)
		}
	} else {
		if s.bestView == nil {
			s.bestView = n
		} else {
			if n.view.GetTipBlock().GetHeight() > s.bestView.view.GetTipBlock().GetHeight() {
				s.bestView = n
			}
			if (n.view.GetTipBlock().GetHeight() == s.bestView.view.GetTipBlock().GetHeight()) && n.view.GetTipBlock().GetTimeslot() < s.bestView.view.GetTipBlock().GetTimeslot() {
				s.bestView = n
			}
		}
	}
}

func (s *ViewGraph) updateConfirmBlock(node *ViewNode) {
	_1block := node.prev
	if _1block == nil {
		s.confirmView = node
		return
	}
	_2block := _1block.prev
	if _2block == nil {
		s.confirmView = _1block
		return
	}
	if _2block.view.GetTimeslot() == _1block.view.GetTimeslot()-1 && _2block.view.GetTimeslot() == node.view.GetTimeslot()-2 {
		s.confirmView = _2block
		return
	}
	s.updateConfirmBlock(_1block)
}
