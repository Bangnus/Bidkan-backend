package entity

import (
	"time"

	"github.com/google/uuid"
)

type IncidentReport struct {
	ID               uuid.UUID  `json:"id"`
	TenantID         uuid.UUID  `json:"tenant_id"`
	RideID           *uuid.UUID `json:"ride_id,omitempty"`
	UserID           *uuid.UUID `json:"user_id,omitempty"`
	ReporterName     string     `json:"reporter_name"`
	ReporterPhone    string     `json:"reporter_phone"`
	IncidentType     string     `json:"incident_type"`     // collision, fall, pedestrian_hit, property_damage, mechanical_failure
	Severity         string     `json:"severity"`          // minor, moderate, severe, critical
	Description      string     `json:"description"`
	LocationName     string     `json:"location_name"`
	LocationLat      float64    `json:"location_lat"`
	LocationLon      float64    `json:"location_lon"`
	PoliceReportNo   string     `json:"police_report_no"`
	InsuranceClaimNo string     `json:"insurance_claim_no"`
	InsuranceStatus  string     `json:"insurance_status"`   // unfiled, filed, under_review, approved, rejected, settled
	ClaimAmount      float64    `json:"claim_amount"`
	ApprovedAmount   float64    `json:"approved_amount"`
	HospitalName     string     `json:"hospital_name"`
	Notes            string     `json:"notes"`
	CreatedAt        time.Time  `json:"created_at"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty"`
}

type IncidentSummary struct {
	TotalIncidents     int               `json:"total_incidents"`
	PendingClaims      int               `json:"pending_claims"`
	TotalClaimedAmount float64           `json:"total_claimed_amount"`
	TotalSettledAmount float64           `json:"total_settled_amount"`
	Incidents          []*IncidentReport `json:"incidents"`
}
