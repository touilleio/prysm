package main

import ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"

type attestationQueue struct {
	queue []*ethpb.IndexedAttestation
}

func newAttestationQueue() *attestationQueue {
	return &attestationQueue{queue: make([]*ethpb.IndexedAttestation, 0)}
}

func (q *attestationQueue) push(att *ethpb.IndexedAttestation) {
	q.queue = append(q.queue, att)
}

func (q *attestationQueue) dequeue() []*ethpb.IndexedAttestation {
	elems := q.queue
	q.queue = make([]*ethpb.IndexedAttestation, 0)
	return elems
}

func (q *attestationQueue) requeue(atts []*ethpb.IndexedAttestation) {
	q.queue = append(q.queue, atts...)
}
