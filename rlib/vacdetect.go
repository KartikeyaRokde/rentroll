package rlib

import (
	"fmt"
	"time"
)

// VacancyMarker is a structure of data defining an increment in time during which a Rentable is vacant
type VacancyMarker struct {
	DtStart time.Time // a period start time
	DtStop  time.Time // end of period
	Amount  float64   // unit market rate during this period
	Comment string    // comment to include with Journal
	State   int64     // Rentable state
}

// VacancyDetect scans the time range specified and looks for pro[rate] periods of time when the
// supplied Rentable is not accounted for. For every period that it is not rented
// a VacancyMarker will be added to an array marking the vacant time period. The return value
// is the list of vacancy markers.
//========================================================================================================
func VacancyDetect(xbiz *XBusiness, d1, d2 *time.Time, r *Rentable) []VacancyMarker {
	var m []VacancyMarker
	var state int64

	//==================================================================
	// Whether it's vacant or not depends on its state. For example,
	// if it is OwnerOccupied, no rent is collected and it is not
	// considered vacant. So, the first thing to do is cache the
	// Rentable state over the period
	//==================================================================
	rsa := GetRentableStatusByRange(r.RID, d1, d2)

	//=====================================================================================
	// In the loop below, we don't want to read the database every iteration
	// for the RTID associated with the rentable as it would result in excessive database
	// reatds. In most cases, the RTID will not change, especially over short periods like
	// a month. So, read all the RTIDs for the period that the loop will process first.
	// then select from them as needed.
	//=====================================================================================
	rta := GetRentableTypeRefsByRange(r.RID, d1, d2) // get the list
	if len(rta) == 0 {
		Ulog("VacancyDetect:  No valid RTID for rentable R%08d during period %s to %s\n",
			r.RID, d1.Format(RRDATEINPFMT), d2.Format(RRDATEINPFMT))
		return m // this is bad! No RTID for the supplied time range
	}
	// rtidMulti := len(rta) > 1 // flag to indicate we need to look for a change in rtid in every pass
	rtid := rta[0].RTID // initialize to the first RTID

	// We may not need to do anything if this rentable is not being managed to budget.  We didn't
	// check it earlier because the code to load the rentable type is here. If there's an issue,
	// just move the code to grabe the RTIDs to the caller and pass the array into this func.
	if xbiz.RT[rtid].ManageToBudget == 0 { // if this rentable is not managing to budget...
		return m // return an empty list now and it will essentially be ignored.
	}

	period := CycleDuration(xbiz.RT[rtid].Proration, *d1)
	t := GetAgreementsForRentable(r.RID, d1, d2) // t is an array of RentalAgreementRentables

	//========================================================
	// Mark vacancy for each time interval between d1 & d2
	//========================================================
	var dtNext time.Time
	k := 0 // number of members of m
	for dt := *d1; dt.Before(*d2); dt = dtNext {
		dtNext = dt.Add(period)
		vacant := true // assume it's vacant and reset if we find it's rented

		// fmt.Printf("VacancyDetect:  %s (%d), period %s - %s\n", r.Name, r.RID, dt.Format(RRDATEINPFMT), dtNext.Format(RRDATEINPFMT))

		rs := SelectRentableStatusForPeriod(&rsa, dt, dtNext)
		state = RENTABLESTATUSONLINE // if there is no state info, we'll assume online
		if len(rs) > 0 {
			state = rs[0].Status // If this turns out to be a problem, maybe we'll choose the state with the greatest percentage of time
		}

		switch state {
		case RENTABLESTATUSONLINE:
			// fmt.Printf("\tonline... ")
			for i := 0; i < len(t); i++ {
				if DateRangeOverlap(&t[i].DtStart, &t[i].DtStop, &dt, &dtNext) {
					// fmt.Printf("covered, RAID = %d\n", t[i].RAID)
					vacant = false // not vacant
				}
			}
		case RENTABLESTATUSADMIN:
			fallthrough
		case RENTABLESTATUSEMPLOYEE:
			fallthrough
		case RENTABLESTATUSOWNEROCC:
			fallthrough
		case RENTABLESTATUSOFFLINE:
			// fmt.Printf("\t{admin|employee|ownerocc|offline}...\n")
		}

		// fmt.Printf("VacancyDetect:  vacant = %v\n", vacant)

		if !vacant {
			continue
		}

		rsa, err := GetRentableSpecialtyTypesForRentableByRange(r, &dt, &dtNext) // this gets an array of rentable specialties that overlap this time period
		if err != nil {
			Ulog("VacancyDetect:  Error retrieving rentable specialties for rentable R%08d during period %s to %s\n",
				r.RID, dt.Format(RRDATEINPFMT), dtNext.Format(RRDATEINPFMT))
			return m // this is bad! No RTID for the supplied time range

		}

		// fmt.Printf("dt = %s, dtNext = %s, r = %s(%d), len(rta) = %d, len(rsa) = %d\n",
		// 	dt.Format("Jan 2"), dtNext.Format("Jan 2"), r.Name, r.RID, len(rta), len(rsa))
		// for iq := 0; iq < len(rta); iq++ {
		// 	fmt.Printf("rta[%d] = (%s - %s) RTID = %d\n", iq, rta[iq].DtStart.Format("1/2/06"), rta[iq].DtStop.Format("1/2/06"), rta[iq].RTID)
		// }

		rentThisPeriod := CalculateGSR(dt, dtNext, r, &rta, rsa, xbiz)
		// fmt.Printf("rentThisPeriod = %8.2f\n", rentThisPeriod)

		//------------------------------------------------
		// optimization to compress consecutive days...
		//------------------------------------------------
		if k > 0 { // If the last entry's DtStop is the same time this one's DtStart...
			if m[k-1].DtStop.Equal(dt) && m[k-1].State == state { // and the umr is at the same rate...
				// m[k-1].Amount += umr * pf // add another increment to the amount
				m[k-1].DtStop = dtNext          // then we'll just adjust the end of that range to include this range too.
				m[k-1].Amount += rentThisPeriod // add the rent for this time increment
				m[k-1].Comment = fmt.Sprintf("vacant %s - %s", m[k-1].DtStart.Format("Jan 2"), m[k-1].DtStop.Format("Jan 2"))
				continue // Range extended.  Next!
			}
		}

		var v VacancyMarker       // ok, this is either the first entry or
		v.DtStart = dt            // it is disjoint from the last range
		v.DtStop = dtNext         // fill it out and
		v.State = state           // note the cause of the vacancy
		v.Amount = rentThisPeriod // save the rate so we don't need to look it up later
		// v.Amount = umr * pf // save the rate so we don't need to look it up later
		v.Comment = fmt.Sprintf("vacant %s - %s", v.DtStart.Format("Jan 2"), v.DtStop.Format("Jan 2"))
		m = append(m, v) // add the new VacancyMarker to the list
		k++
	}
	return m
}
