package Transition

import (
	. "chukuparser/nlp/types"
	"reflect"
	"sort"
	"strings"
)

type StackArray struct {
	Array []int
}

var _ Stack = &StackArray{}

func (s *StackArray) Equal(other Stack) bool {
	return reflect.DeepEqual(s, other)
}

func (s *StackArray) Clear() {
	s.Array = s.Array[0:0]
}

func (s *StackArray) Push(val int) {
	s.Array = append(s.Array, val)
}

func (s *StackArray) Pop() (int, bool) {
	if s.Size() == 0 {
		return 0, false
	}
	retval := s.Array[len(s.Array)-1]
	s.Array = s.Array[:len(s.Array)-1]
	return retval, true
}

func (s *StackArray) Index(index int) (int, bool) {
	if index >= s.Size() {
		return 0, false
	}
	return s.Array[len(s.Array)-1-index], true
}

func (s *StackArray) Peek() (int, bool) {
	result, exists := s.Index(0)
	return result, exists
}

func (s *StackArray) Size() int {
	return len(s.Array)
}

func (s *StackArray) Copy() Stack {
	newArray := make([]int, len(s.Array), cap(s.Array))
	copy(newArray, s.Array)
	newStack := Stack(&StackArray{newArray})
	return newStack
}

func NewStackArray(size int) *StackArray {
	return &StackArray{make([]int, 0, size)}
}

// type QueueSlice struct {
// 	slice       []int
// 	hasDequeued bool
// }

// func (q *QueueSlice) Clear() {
// 	q.slice = q.slice[0:0]
// }

// func (q *QueueSlice) Enqueue(val int) {
// 	if q.hasDequeued {
// 		panic("Can't Enqueue after Dequeue")
// 	}
// 	q.slice = append(q.slice, val)
// }

// func (q *QueueSlice) Dequeue() (int, bool) {
// 	if q.Size() == 0 {
// 		return 0, false
// 	}
// 	retval := q.slice[0]
// 	return retval, true
// }

// func (q *QueueSlice) Index(index int) (int, bool) {
// 	if index >= q.Size() {
// 		return 0, false
// 	}
// 	return q.slice[index], true
// }

// func (q *QueueSlice) Peek() (int, bool) {
// 	result, exists := q.Index(0)
// 	return result, exists
// }

// func (q *QueueSlice) Size() int {
// 	return len(q.slice)
// }

// func (q *QueueSlice) Copy() QueueSlice {
// 	return QueueSlice{q.slice, q.hasDequeued}
// }

// func NewQueueSlice(size int) QueueSlice {
// 	return QueueSlice{make([]int, 0, size), false}
// }

type ArcSetSimple struct {
	Arcs         []LabeledDepArc
	SeenHead     map[int]bool
	SeenModifier map[int]bool
	SeenArc      map[[2]int]bool
}

var _ ArcSet = &ArcSetSimple{}
var _ sort.Interface = &ArcSetSimple{}

func (s *ArcSetSimple) Less(i, j int) bool {
	if s.Arcs[i].GetHead() < s.Arcs[j].GetHead() {
		return true
	}
	if s.Arcs[i].GetHead() == s.Arcs[j].GetHead() {
		return s.Arcs[i].GetModifier() < s.Arcs[j].GetModifier()
	}
	return false
}

func (s *ArcSetSimple) Swap(i, j int) {
	s.Arcs[i], s.Arcs[j] = s.Arcs[j], s.Arcs[i]
}

func (s *ArcSetSimple) Len() int {
	return s.Size()
}

func (s *ArcSetSimple) ValueComp(i, j int, other *ArcSetSimple) int {
	left := s.Arcs[i]
	right := other.Arcs[j]
	if reflect.DeepEqual(left, right) {
		return 0
	}
	if left.GetModifier() < right.GetModifier() {
		return 1
	}
	return -1
}

func (s *ArcSetSimple) Equal(other ArcSet) bool {
	if s.Size() == 0 && other.Size() == 0 {
		return true
	}
	copyThis := s.Copy().(*ArcSetSimple)
	copyOther := other.Copy().(*ArcSetSimple)
	if copyThis.Len() != copyOther.Len() {
		return false
	}
	sort.Sort(copyThis)
	sort.Sort(copyOther)
	for i, _ := range copyThis.Arcs {
		if !copyThis.Arcs[i].Equal(copyOther.Arcs[i]) {
			return false
		}
	}
	return true
}

func (s *ArcSetSimple) Sorted() *ArcSetSimple {
	copyThis := s.Copy().(*ArcSetSimple)
	sort.Sort(copyThis)
	return copyThis
}

func (s *ArcSetSimple) Diff(other ArcSet) (ArcSet, ArcSet) {
	copyThis := s.Copy().(*ArcSetSimple)
	copyOther := other.Copy().(*ArcSetSimple)
	sort.Sort(copyThis)
	sort.Sort(copyOther)

	leftOnly := NewArcSetSimple(copyThis.Len())
	rightOnly := NewArcSetSimple(copyOther.Len())
	i, j := 0, 0
	for i < copyThis.Len() && j < copyOther.Len() {
		comp := copyThis.ValueComp(i, j, copyOther)
		switch {
		case comp == 0:
			i++
			j++
		case comp < 0:
			leftOnly.Add(copyThis.Arcs[i])
			i++
		case comp > 0:
			rightOnly.Add(copyOther.Arcs[j])
			j++
		}
	}
	return leftOnly, rightOnly
}

func (s *ArcSetSimple) Copy() ArcSet {
	newArcs := make([]LabeledDepArc, len(s.Arcs), cap(s.Arcs))
	// headMap, modMap, arcMap := make(map[int]bool, cap(s.Arcs)), make(map[int]bool, cap(s.Arcs)), make(map[[2]int]bool, cap(s.Arcs))
	// for k, v := range s.SeenArc {
	// 	arcMap[k] = v
	// }
	// for k, v := range s.SeenHead {
	// 	headMap[k] = v
	// }
	// for k, v := range s.SeenModifier {
	// 	modMap[k] = v
	// }
	copy(newArcs, s.Arcs)
	return ArcSet(&ArcSetSimple{Arcs: newArcs})
	// return ArcSet(&ArcSetSimple{newArcs, headMap, modMap, arcMap})
}

func (s *ArcSetSimple) Clear() {
	s.Arcs = s.Arcs[0:0]
}

func (s *ArcSetSimple) Index(i int) LabeledDepArc {
	if i >= len(s.Arcs) {
		return nil
	}
	return s.Arcs[i]
}

func (s *ArcSetSimple) Add(arc LabeledDepArc) {
	// s.SeenHead[arc.GetHead()] = true
	// s.SeenModifier[arc.GetModifier()] = true
	// s.SeenArc[[2]int{arc.GetHead(), arc.GetModifier()}] = true
	s.Arcs = append(s.Arcs, arc)
}

func (s *ArcSetSimple) Get(query LabeledDepArc) []LabeledDepArc {
	var results []LabeledDepArc
	head := query.GetHead()
	modifier := query.GetModifier()
	relation := query.GetRelation()
	for _, arc := range s.Arcs {
		if head >= 0 && head != arc.GetHead() {
			continue
		}
		if modifier >= 0 && modifier != arc.GetModifier() {
			continue
		}
		if string(relation) != "" && relation != arc.GetRelation() {
			continue
		}
		results = append(results, arc)
	}
	return results
}

func (s *ArcSetSimple) Size() int {
	return len(s.Arcs)
}

func (s *ArcSetSimple) Last() LabeledDepArc {
	if s.Size() == 0 {
		return nil
	}
	return s.Arcs[len(s.Arcs)-1]
}

func (s *ArcSetSimple) String() string {
	arcs := make([]string, s.Size())
	for i, arc := range s.Arcs {
		arcs[i] = arc.String()
	}
	return strings.Join(arcs, "\n")
}

func (s *ArcSetSimple) HasHead(modifier int) bool {
	// _, exists := s.SeenModifier[modifier]
	// return exists
	return len(s.Get(&BasicDepArc{-1, -1, modifier, DepRel("")})) > 0
}

func (s *ArcSetSimple) HasModifiers(head int) bool {
	// _, exists := s.SeenHead[head]
	// return exists
	return len(s.Get(&BasicDepArc{head, -1, -1, DepRel("")})) > 0
}

func (s *ArcSetSimple) HasArc(head, modifier int) bool {
	_, exists := s.SeenArc[[2]int{head, modifier}]
	return exists
}

func NewArcSetSimple(size int) *ArcSetSimple {
	return &ArcSetSimple{
		Arcs: make([]LabeledDepArc, 0, size),
		// SeenHead:     make(map[int]bool, size),
		// SeenModifier: make(map[int]bool, size),
		// SeenArc:      make(map[[2]int]bool, size),
	}
}

func NewArcSetSimpleFromGraph(graph LabeledDependencyGraph) *ArcSetSimple {
	arcSet := NewArcSetSimple(graph.NumberOfEdges())
	for _, edgeNum := range graph.GetEdges() {
		arc := graph.GetLabeledArc(edgeNum)
		arcSet.Add(arc)
	}
	return arcSet
}
