package slasher

import "time"

func (s *Service) detectIncomingAttestations() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Info("Attestation received by slasher")
		}
	}
}
