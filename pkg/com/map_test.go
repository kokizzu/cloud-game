package com

import "testing"

func TestMap_Base(t *testing.T) {
	// map map
	m := Map[int, int]{m: make(map[int]int)}

	if m.Len() > 0 {
		t.Errorf("should be empty, %v %v", m.Len(), m.m)
	}
	k := 0
	m.Put(k, 0)
	if m.Len() == 0 {
		t.Errorf("should not be empty, %v", m.m)
	}
	if !m.Has(k) {
		t.Errorf("should have the key %v, %v", k, m.m)
	}
	v, ok := m.Get(k)
	if v != 0 && !ok {
		t.Errorf("should have the key %v and ok, %v %v", k, ok, m.m)
	}
	_, ok = m.Get(k + 1)
	if ok {
		t.Errorf("should not find anything, %v %v", ok, m.m)
	}
	m.Put(1, 1)
	v, ok = m.FindBy(func(v int) bool { return v == 1 })
	if v != 1 && !ok {
		t.Errorf("should have the key %v and ok, %v %v", 1, ok, m.m)
	}
	sum := 0
	for v := range m.Values() {
		sum += v
	}
	if sum != 1 {
		t.Errorf("shoud have exact sum of 1, but have %v", sum)
	}
	m.Remove(1)
	if !m.Has(0) || m.Len() > 1 {
		t.Errorf("should remove only one element, but has %v", m.m)
	}
	m.Put(3, 3)
	v = m.Pop(3)
	if v != 3 {
		t.Errorf("should have value %v, but has %v %v", 3, v, m.m)
	}
	m.Remove(3)
	m.Remove(0)
	if m.Len() != 0 {
		t.Errorf("should be completely empty, but %v", m.m)
	}
}

func TestMap_Remove(t *testing.T) {
	m := Map[int, int]{m: make(map[int]int)}

	m.Put(1, 10)
	m.Put(2, 20)

	if m.Len() == 0 {
		t.Error("should not be empty after puts")
	}

	m.Remove(1)
	if m.Len() == 0 {
		t.Error("should not be empty, key 2 remains")
	}
	if m.Len() != 1 {
		t.Errorf("expected len 1, got %v", m.Len())
	}

	m.Remove(2)
	if m.Len() != 0 {
		t.Error("should be empty after removing last key")
	}

	// Remove of non-existent key is a no-op.
	m.Remove(99)
	if m.Len() != 0 {
		t.Error("should still be empty after no-op remove")
	}
}

func TestMap_Concurrency(t *testing.T) {
	m := Map[int, int]{m: make(map[int]int)}
	for i := range 100 {
		go m.Put(i, i)
		go m.Has(i)
		go m.Pop(i)
	}
}
