package gtree

import (
	"fmt"
	"testing"
)

func makePerson(id int, name string) *FanPerson {
	return &FanPerson{
		ID:       id,
		Headings: []string{name},
		Details:  []string{},
	}
}

func buildGenerations(n int, nextID *int) *FanPerson {
	if n <= 0 {
		return nil
	}
	p := makePerson(*nextID, fmt.Sprintf("Gen%d, id%d", n, *nextID))
	*nextID++
	if n == 1 {
		return p
	}
	p.Father = buildGenerations(n-1, nextID)
	p.Mother = buildGenerations(n-1, nextID)
	return p
}

func TestAssignFanNodes_Generations(t *testing.T) {
	// const tolerance = 1e-6

	// testCases := []struct {
	// 	name       string
	// 	gens       int
	// 	expectSize int // number of persons expected in full binary tree
	// }{
	// 	{"1 generation", 1, 1},
	// 	{"2 generations", 2, 3},  // 1 + 2
	// 	{"3 generations", 3, 7},  // 1 + 2 + 4
	// 	{"4 generations", 4, 15}, // 1 + 2 + 4 + 8
	// }

	// for _, tc := range testCases {
	// 	t.Run(tc.name, func(t *testing.T) {
	// 		id := 1
	// 		root := buildGenerations(tc.gens, &id)

	// 		opts := &FanLayoutOptions{
	// 			MaxGenerations: tc.gens,
	// 			RadiusStep:     100,
	// 			BaseFontSize:   12,
	// 			FontScale:      0.9,
	// 			AngularSpanDeg: 120,
	// 		}

	// 		nodes, err := assignFanNodes(root, opts)
	// 		if err != nil {
	// 			t.Errorf("unexpected error from assignFanNodes: %v", err)
	// 			t.FailNow()
	// 		}

	// 		if len(nodes) != tc.expectSize {
	// 			t.Errorf("expected %d nodes, got %d", tc.expectSize, len(nodes))
	// 		}

	// 		// Prepare expected layout geometry
	// 		totalSlots := 1 << uint(opts.MaxGenerations)
	// 		slotAngle := degreesToRadians(opts.AngularSpanDeg) / float64(totalSlots)
	// 		startAngle := -degreesToRadians(opts.AngularSpanDeg) / 2

	// 		// Track number of persons per generation and position
	// 		genCounts := make(map[int]int)

	// 		for _, node := range nodes {
	// 			p := polarToCartesian(node.Radius, node.Angle)
	// 			t.Logf("person %d: gen=%d, radius=%g, angle=%.2f, x=%.2f, y=%.2f", node.Person.ID, node.Gen, node.Radius, radiansToDegrees(node.Angle), x, y)

	// 			if node.Gen < 0 || node.Gen > tc.gens {
	// 				t.Errorf("invalid generation %d for person ID %d", node.Gen, node.Person.ID)
	// 			}
	// 			if node.Radius != float64(node.Gen)*opts.RadiusStep {
	// 				t.Errorf("wrong radius: got %.2f, expected %.2f",
	// 					node.Radius, float64(node.Gen)*opts.RadiusStep)
	// 			}
	// 			if math.IsNaN(node.Angle) || math.IsInf(node.Angle, 0) {
	// 				t.Errorf("invalid angle: %.2f", node.Angle)
	// 			}

	// 			expectedRadius := float64(node.Gen) * opts.RadiusStep
	// 			if math.Abs(node.Radius-expectedRadius) > tolerance {
	// 				t.Errorf("person %d: got radius %.4f, want %.4f", node.Person.ID, node.Radius, expectedRadius)
	// 			}

	// 			genSlots := 1 << uint(node.Gen)
	// 			offset := (totalSlots - genSlots) / 2
	// 			slot := offset + genCounts[node.Gen]
	// 			genCounts[node.Gen]++

	// 			expectedAngle := startAngle + float64(slot)*slotAngle + slotAngle/2
	// 			// if node.Side != 0 {
	// 			// 	expectedAngle += float64(node.Side) * slotAngle * 0.2
	// 			// }

	// 			if math.Abs(node.Angle-expectedAngle) > tolerance {
	// 				t.Errorf("person %d: got angle %.6f, want %.6f", node.Person.ID, node.Angle, expectedAngle)
	// 			}

	// 		}
	// 	})
	// }
}
