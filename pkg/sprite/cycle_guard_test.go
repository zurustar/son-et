package sprite

import "testing"

// TestAddChildRejectsCycles is a regression test for the bug where AddChild had
// no cycle guard, allowing self-parenting or mutual parent/child links that made
// recursive traversals (AbsolutePosition, drawSprite, ...) stack-overflow.
// See docs/bug-hunt-findings.md finding F.
func TestAddChildRejectsCycles(t *testing.T) {
	// Self cycle must be rejected.
	a := NewSprite(1, nil)
	a.AddChild(a)
	if a.parent == a {
		t.Errorf("self-parenting was allowed (a.parent == a)")
	}
	if len(a.children) != 0 {
		t.Errorf("self was added as its own child: children=%d", len(a.children))
	}

	// Mutual cycle must be rejected.
	b := NewSprite(2, nil)
	c := NewSprite(3, nil)
	b.AddChild(c)
	c.AddChild(b) // would create b <-> c cycle
	if b.parent == c && c.parent == b {
		t.Errorf("mutual parent/child cycle was allowed (b <-> c)")
	}

	// Deeper ancestor cycle: a1 -> a2 -> a3, then a3.AddChild(a1) must be rejected.
	a1 := NewSprite(10, nil)
	a2 := NewSprite(11, nil)
	a3 := NewSprite(12, nil)
	a1.AddChild(a2)
	a2.AddChild(a3)
	a3.AddChild(a1) // a1 is an ancestor of a3 -> cycle
	if a1.parent == a3 {
		t.Errorf("ancestor cycle was allowed (a1.parent == a3)")
	}

	// Sanity: a legitimate parent/child relationship still works, and traversal
	// terminates (would hang/overflow if a cycle slipped through).
	p := NewSprite(20, nil)
	ch := NewSprite(21, nil)
	p.SetPosition(10, 20)
	ch.SetPosition(3, 4)
	p.AddChild(ch)
	if ch.parent != p {
		t.Fatalf("legitimate AddChild failed: ch.parent = %v", ch.parent)
	}
	if x, y := ch.AbsolutePosition(); x != 13 || y != 24 {
		t.Errorf("AbsolutePosition = (%v,%v), want (13,24)", x, y)
	}
}
