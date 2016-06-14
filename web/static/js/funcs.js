// global map airport->city.
airpCityMap = new Map([["AEP", "Ciudad de Buenos Aires"],["EZE","Ciudad de Buenos Aires"],["BHI","Bahía Blanca"],["BRC","San Carlos de Bariloche"],["CTC","San Fernando del Valle de Catamarca"],["CRD","Comodoro Rivadavia"],["COR","Córdoba"],["CNQ","Corrientes"],["FTE","El Calafate"],["EQS","Esquel"],["FMA","Formosa"],["IGR","Puerto Iguazú"],["JUJ","Jujuy"],["IRJ","La Rioja"],["MDQ","Mar del Plata"],["MDZ","Mendoza"],["NQN","Neuquén"],["PRA","Paraná"],["PSS","Posadas"],["RES","Resistencia"],["RGL","Río Gallegos"],["RGA","Río Grande"],["RHD","Termas de Río Hondo"],["ROS","Rosario"],["SLA","Salta"],["UAQ","San Juan"],["LUQ","San Luís"],["CPC","San Martín de los Andes"],["AFA","San Rafael"],["SFN","Santa Fe"],["RSA","Santa Rosa"],["SDE","Santiago del Estero"],["REL","Trelew"],["TUC","San Miguel de Tucumán"],["USH","Ushuaia"],["VDM","Viedma"]])

// current page number
pageNumber = 1

depYear = 0

// filter sliders have changed
priceFilterChanged = false
depHourFilterChanged = false
arrDayFilterDays = [-1]
transfersFilterChanged = false

// index for checking if post result is consistent with the current search
queryInd = 0

// check if first combination json has arrived
firstCombJsonArrived = false
firstCombJson = {}
combJsonInd = 0

allTripsJson = []


function swapCities() {
	var origVal = $('#select-orig-city').selectize()[0].selectize.getValue()
	var destVal = $('#select-dest-city').selectize()[0].selectize.getValue()
	$('#select-orig-city').selectize()[0].selectize.setValue(destVal, true)
	$('#select-dest-city').selectize()[0].selectize.setValue(origVal, true)
}

function getTripTotalPrice(trip) {
	//return parseInt(trip.children[1].children[0].children[0].textContent.slice(0, -1).replace(/\./g, ''))
	var jsonTrips = allTripsJson[trip.id]
	var total = 0.0
	for (i = 0; i < jsonTrips.length; i++) {
		total += jsonTrips[i].TotalPrice
	}
	return Math.round(total)
}

function getTripAdultPrice(trip) {
	// return parseInt(trip.children[1].children[0].children[1].textContent.split("$")[0].substring(1).replace(/\./g, ''))
	var jsonTrips = allTripsJson[trip.id]
	var total = 0.0
	for (i = 0; i < jsonTrips.length; i++) {
		total += jsonTrips[i].PricePerAdult
	}
	return Math.round(total)
}

function getTripDepHourMin(trip) {
	// var hourMin = trip.children[0].children[1].children[0].children[0].children[0].textContent.split(":")
	// var hour = parseInt(hourMin[0])
	// var min = parseInt(hourMin[1])
	// return [hour, min]
	var jsonTrips = allTripsJson[trip.id]
	return [jsonTrips[0].DepHour, jsonTrips[0].DepMin]
}

function getTripArrDayMonth(trip) {
	// dayMonth = trip.children[0].children[1].children[2].children[0].children[2].textContent.split("/")
	// var day = parseInt(dayMonth[0])
	// var month = parseInt(dayMonth[1])
	// return [day, month]
	var jsonTrips = allTripsJson[trip.id]
	return [jsonTrips[jsonTrips.length-1].ArrDay, jsonTrips[jsonTrips.length-1].ArrMonth]
}

function getTripDepDayMonth(trip) {
	// var dayMonth = trip.children[0].children[1].children[0].children[0].children[2].textContent.split("/")
	// var day = parseInt(dayMonth[0])
	// var month = parseInt(dayMonth[1])
	// return [day, month]
	var jsonTrips = allTripsJson[trip.id]
	return [jsonTrips[0].DepDay, jsonTrips[0].DepMonth]
}

function getTravelDays(trip) {
	var dep = getTripDepDayMonth(trip); var depDay = dep[0]; var depMonth = dep[1]
 	var arr = getTripArrDayMonth(trip); var arrDay = arr[0]; var arrMonth = arr[1]
 	var arrYear = depYear; if (depMonth == 12 && arrMonth == 1) { arrYear++ }
 	var depDate = new Date(depYear, depMonth-1, depDay)
 	var arrDate = new Date(arrYear, arrMonth-1, arrDay) 
 	var diffDays = Math.ceil((arrDate - depDate) / (1000 * 3600 * 24));
 	return diffDays
}

function getTripTransfers(trip) {
	// return Math.floor(trip.children[0].children[0].children[0].children[0].children[0].children.length / 2)
	var jsonTrips = allTripsJson[trip.id]
	return jsonTrips.length-1
}

function buscar() {

	// error checking

	// cities
	var idOrig = document.getElementById('select-orig-city').value;
	var idDest = document.getElementById('select-dest-city').value;

	var alerts = ''

	if (idOrig == '' || idDest == '' ) {
		alerts += 
			`<div class="col-lg-11"> 
                <div class="col-lg-11 col-lg-offset-1" style="padding:0px">
                  	<div class="alert-style alert alert-warning fade in">
                  		<div class="alert-content">
							<a class="close" data-dismiss="alert" href="#">&times;</a>
							<p>Che, poneme de dónde querés salir y a dónde querés llegar, por favor!</p>
						</div>
					</div>
                </div>
			</div>`
	} else if (idOrig == idDest) {
		 alerts += 
			`<div class="col-lg-11"> 
                <div class="col-lg-11 col-lg-offset-1" style="padding:0px">
                  	<div class="alert-style alert alert-warning fade in">
                  		<div class="alert-content">
							<a class="close" data-dismiss="alert" href="#">&times;</a>
							<p>Che, poné una ciudad de llegada distinta a la de salida!</p>
						</div>
					</div>
                </div>
			</div>`
	}

	var fecha = document.getElementById('dep-date').value;
	
	if (fecha == '') {
		alerts += 
			`<div class="col-lg-11"> 
                <div class="col-lg-11 col-lg-offset-1" style="padding:0px">
                  	<div class="alert-style alert alert-warning fade in">
                  		<div class="alert-content">
							<a class="close" data-dismiss="alert" href="#">&times;</a>
							<p>Fijate que no pusiste la fecha de salida. Así no puedo trabajar!</p>
						</div>
					</div>
                </div>
			</div>`
	}

	document.getElementById('alerts').innerHTML = alerts

	if (alerts == '') {
		getTrips();
	}

}

function showErrorChildren() {
    children2 = parseInt($('.popover-content #children2').val())
  infants = parseInt($('.popover-content #infants').val())
  adults = parseInt($("#adults").val())
  if ( children2 + infants > adults ) {
    // show alert
    $('.popover-content #children2').attr("style", "background-color: papayawhip");
    $('.popover-content #infants').attr("style", "background-color: papayawhip");
    document.getElementById("alert-children").innerHTML = '<p class="text-warning" style="font-size: 12px; text-align:justify"><small>Viajando en micro, todos los menores de 5 años deben ser acompañados por un adulto. No se tendrán en cuenta los viajes en micro de esta manera.</small></p>'
  } else {
    // hide alert
    $('.popover-content #children2').attr("style", "background-color: white");
    $('.popover-content #infants').attr("style", "background-color: white");
    document.getElementById("alert-children").innerHTML = ''
  }
}

function getFormInfo() {
	adults = $("#adults").val()
	children11 = $("#children5").val()
	children5 = $("#children2").val()
	infants = $("#infants").val()
	var depdate = $("#dep-date").val()
	var depcity = document.getElementById("select-orig-city")
	dep = depcity.options[depcity.selectedIndex].value
	depName = depcity.options[depcity.selectedIndex].text.split("(")[0].trim()
	var arrcity = document.getElementById("select-dest-city")
	arr = arrcity.options[arrcity.selectedIndex].value
	arrName = arrcity.options[arrcity.selectedIndex].text.split("(")[0].trim()
	var yearmonthday = depdate.split("/")
	day = yearmonthday[0]
	month = yearmonthday[1]
	year = yearmonthday[2]
}


function getTrips() {

	// gather information

	getFormInfo()

	// connect to flights

	window.location.href = '/flights?adults=' + adults + 
											'&children11='+children11 +
											'&children5='+children5 + 
											'&infants='+infants + 
											'&dep='+dep + 
											'&depName='+depName + 
											'&arr='+arr + 
											'&arrName='+arrName + 
											'&year='+year + 
											'&month='+month + 
											'&day='+day

}

function isValidTrip(trip) {

	// check price range
	var slider = $('#price-filter')
	var minPrice = slider.slider('getAttribute', 'min')
	var maxPrice = slider.slider('getValue')
	var price = getTripTotalPrice(trip)

	if (price < minPrice || price > maxPrice) {
		return false
	}

	// check dep hour range
	slider = $('#dep-hour-filter')
	var depTime = slider.slider('getValue')
	var minDepTime = depTime[0]
	var maxDepTime = depTime[1]
	var depHourMin = getTripDepHourMin(trip)
	var time = depHourMin[0]*60+depHourMin[1]

	if (time < minDepTime || time > maxDepTime) {
		return false
	}

	// check travel days
	var travelDays = getTravelDays(trip)
	var checked = document.querySelector('input[name="arr-day"]:checked').value;
	if (checked != -1 && travelDays > checked) {
		return false
	}

	// check max transfers
	var transfers = getTripTransfers(trip)
	var maxTransf = $('#transfers-filter').slider('getValue')
	if (transfers > maxTransf) {
		return false
	}


	return true

}

function updatePagination() {

	var trips = document.getElementById('trips').children 

	// find current trips number, taking filters into account
	var tripsNumber = 0
	for (var i = 0; i < trips.length; i++) {
		if (isValidTrip(trips[i])) {
			tripsNumber++
		}
	}

	// update "results"
	$('#number-of-results').html(tripsNumber + ' resultados')

	//  update page numbers
	var maxPages = Math.max(Math.ceil(tripsNumber/10.0),1)
	if (maxPages < pageNumber) {
		pageNumber = maxPages
	}

	var validTrips = 0
	for (var i = 0; i < trips.length; i++) {
		if (!isValidTrip(trips[i])) {
			trips[i].style.display = "none"
		} else {
			var start = (pageNumber-1)*10
			var end = start + 10
			if (validTrips >= start && validTrips < end) {
				trips[i].style.display = "block"
			} else {
				trips[i].style.display = "none"
			}
			validTrips++
		}
	}

	// update page numbers
	$('.pages').bootpag({total: Math.max(Math.ceil(validTrips/10.0), 1)});
}

function updatePriceFilter(minPrice, maxPrice) {
	var slider = $('#price-filter')
	var min = Math.floor(minPrice)
	var max = Math.floor(maxPrice)
	slider.slider('setAttribute','min', min)
	slider.slider('setAttribute','max', max)
	$('#min-price').html('<strong>'+addPoints(min)+'$</strong>')
	if (!priceFilterChanged) { 
		$('#max-price').html('<strong>'+ addPoints(max) +'$</strong>')
		 slider.slider('setValue', max)
	}
}

function updateDepHourFilter(minTime, maxTime) {
	var slider = $('#dep-hour-filter')
	var minHour = Math.floor(minTime/60); var minMin = minTime%60
	var maxHour = Math.floor(maxTime/60); var maxMin = maxTime%60
	slider.slider('setAttribute','min', minTime)
	slider.slider('setAttribute','max', maxTime)
	if (!depHourFilterChanged) {
		$('#min-dep-hour').html('<strong>'+("0" + minHour).slice(-2)+':'+("0" + minMin).slice(-2)+'</strong>')
		$('#max-dep-hour').html('<strong>'+("0" + maxHour).slice(-2)+':'+("0" + maxMin).slice(-2)+'</strong>')
		slider.slider('setValue', [minTime, maxTime])
	}
}

function updateArrDayFilter() {
	var sorting = function(a,b) {
		return a-b
	}
	arrDayFilterDays.sort(sorting)
	var checked = document.querySelector('input[name="arr-day"]:checked').value;
	var html = ''
	for (var i=0; i < arrDayFilterDays.length; i++) {
		var num = arrDayFilterDays[i]
		var checkedHtml = ''
		if (num == checked) {
			checkedHtml = 'checked="checked"'
		}
		if (num == -1) {
			html += '<label><input type="radio" value="-1" name="arr-day"' + checkedHtml + ' onclick="updatePagination()"><span>Cuando sea</span></label>'
		} else if (num == 0) {
			html += '<label><input type="radio" value="0" name="arr-day"' + checkedHtml + ' onclick="updatePagination()"><span>El mismo día</span></label>'
		} else if (num == 1) {
			html += '<label><input type="radio" value="1" name="arr-day"' + checkedHtml + ' onclick="updatePagination()"><span>En 1 día o menos</span></label>'
		} else {
			html += '<label><input type="radio" value="'+num+'" name="arr-day"' + checkedHtml + ' onclick="updatePagination()"><span>En '+num+' días o menos</span></label>'
		}
	}
	$('#arr-day-filter').html(html)
}

function updateTransfersFilter(maxTransf) {
	var slider = $('#transfers-filter')
	slider.slider('setAttribute','max', maxTransf)
	if (!transfersFilterChanged) {
		slider.slider('setValue', maxTransf)
		$('#max-transfers').html(maxTransf)
	}
}

function addPoints(n){
    var rx=  /(\d+)(\d{3})/;
    return String(n).replace(/^\d+/, function(w){
        while(rx.test(w)){
            w= w.replace(rx, '$1.$2');
        }
        return w;
    });
}

function  sort_by_total_price(a, b) {
	var aPrice = getTripTotalPrice(a)
	var bPrice = getTripTotalPrice(b)
	return aPrice - bPrice
}

function sort_by_adult_price(a, b) {
	var aPrice = getTripAdultPrice(a)
	var bPrice = getTripAdultPrice(b)
	return aPrice - bPrice
}

function sort_by_dep_time(a, b) {
	var hourMin = getTripDepHourMin(a)
	var aHour = parseInt(hourMin[0])
	var aMin = parseInt(hourMin[1])
	var dayMonth = getTripDepDayMonth(a)
	var aDay = parseInt(dayMonth[0])
	var aMonth = parseInt(dayMonth[1])

	hourMin = getTripDepHourMin(b)
	var bHour = parseInt(hourMin[0])
	var bMin = parseInt(hourMin[1])
	dayMonth = getTripDepDayMonth(b)
	var bDay = parseInt(dayMonth[0])
	var bMonth = parseInt(dayMonth[1])

	if (aMonth != bMonth) {
		if (aMonth == 12 && bMonth == 1) {
			return -1
		} else {
			return aMonth - bMonth
		}
	} else {
		if (aDay != bDay) {
			return aDay - bDay
		} else {
			if (aHour != bHour) {
				return aHour - bHour
			} else if (aMin != bMin) {
				return aMin - bMin
			} else {
				return sort_by_total_price(a,b)
			}
		}
	}
}

function sort_by_arr_time(a, b) {
	var hourMin = a.children[0].children[1].children[2].children[0].children[0].textContent.split(":")
	var aHour = parseInt(hourMin[0])
	var aMin = parseInt(hourMin[1])
	var dayMonth = getTripArrDayMonth(a)
	var aDay = dayMonth[0]
	var aMonth = dayMonth[1]

	hourMin = b.children[0].children[1].children[2].children[0].children[0].textContent.split(":")
	var bHour = parseInt(hourMin[0])
	var bMin = parseInt(hourMin[1])
	dayMonth = getTripArrDayMonth(b)
	var bDay = dayMonth[0]
	var bMonth = dayMonth[1]

	if (aMonth != bMonth) {
		if (aMonth == 12 && bMonth == 1) {
			return -1
		} else {
			return aMonth - bMonth
		}
	} else {
		if (aDay != bDay) {
			return aDay - bDay
		} else {
			if (aHour != bHour) {
				return aHour - bHour
			} else if (aMin != bMin) {
				return aMin - bMin
			} else {
				return sort_by_total_price(a,b)
			}
		}
	}
}

function sort_by_transfers(a, b) {
	var aTransf = getTripTransfers(a)
	var bTransf = getTripTransfers(b)
	if (aTransf == bTransf) {
		return sort_by_total_price(a,b)
	}
	return aTransf - bTransf
}

function sortTrips(allHtmlTripsObj) {
	// get sorting method
	var val = $('#sorting').val()
	if (val == 0) {
		allHtmlTripsObj.html.sort(sort_by_total_price)
	} else if (val == 1) {
		allHtmlTripsObj.html.sort(sort_by_adult_price)
	} else if (val == 2) {
		allHtmlTripsObj.html.sort(sort_by_dep_time)
	} else if (val == 3) {
		allHtmlTripsObj.html.sort(sort_by_arr_time)
	} else if (val == 4) {
		allHtmlTripsObj.html.sort(sort_by_transfers)
	}
	
}

function sortAgain() {
	var divs = document.getElementById('trips').children
	var allHtmlTrips = []
	for (var i = 0; i < divs.length; i++) {
		allHtmlTrips.push(divs[i])
	}
	// sort the trips
	var obj = {html: allHtmlTrips}
	sortTrips(obj)
	allHtmlTrips = obj.html
	
	// insert
	for (var i = 0; i < allHtmlTrips.length; i++) {
        document.getElementById('trips').appendChild(allHtmlTrips[i]);
    }

    updatePagination()
}



// First, generates the html code for the newly arrived trips (jsonResult). Then, concatenates them with the existing trips in the DOM,
// and refreshes the list by pushing every element in the concatenated and sorted list to the DOM.
function addCombinationsSorted(json, depName, arrName) {

	if (json == null) {
		// no results
		return
	}
	
	var tripId = allTripsJson.length
	allTripsJson = allTripsJson.concat(json)

	var htmlCurrentTrips = document.getElementById('trips').children
	var newTrips = ''

	for (i = 0; i < json.length; i++, tripId++) {
		var trips = json[i]

		totalPrice = 0; pricePerAdult = 0;

		depTrip = trips[0]
		depHour = depTrip.DepHour
		depMin = depTrip.DepMin
		depDay = depTrip.DepDay
		depMonth = depTrip.DepMonth

		arrTrip = trips[trips.length-1]
		arrHour = arrTrip.ArrHour
		arrMin = arrTrip.ArrMin
		arrDay = arrTrip.ArrDay
		arrMonth = arrTrip.ArrMonth

		var transpTypes = [] // 0 -> plane, 1 -> bus

		modal = '<div class="modal fade" id="myModal'+tripId+'" tabindex="-1" role="dialog" aria-labelledby="myModalLabel">\
                  <div class="modal-dialog" role="document">\
                    <div class="modal-content">\
                      <div class="modal-header">\
                        <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>\
                        <h4 class="modal-title" id="myModalLabel"><strong>'+depName+'</strong> a <strong>'+arrName+'</strong></h4>\
                      </div>\
                      <div class="modal-body">'

		for (j = 0; j < trips.length; j++) {

			trip = trips[j]

			price = Math.round(trip.TotalPrice)
			priceAdult = Math.round(trip.PricePerAdult)
			totalPrice += price
			pricePerAdult += priceAdult

			depTime = ("0" + trip.DepHour).slice(-2) + ':' + ("0" + trip.DepMin).slice(-2)
			arrTime = ("0" + trip.ArrHour).slice(-2) + ':' + ("0" + trip.ArrMin).slice(-2)
			plusOneDay = ''
			if (trip.ArrHour < trip.DepHour) {
				plusOneDay = ' <small>(+1)</small>'
			}

			if (trip.DepAirp.length == 3) {
				// airport
				transpTypes.push(0)
				pngTransp = 'airplane-flight.png'

				depAirp = trip.DepAirp
				depCity = airpCityMap.get(depAirp).split("(")[0].trim()
				arrAirp = trip.ArrAirp
				arrCity = airpCityMap.get(arrAirp).split("(")[0].trim()

			} else {
				transpTypes.push(1)
				pngTransp = 'bus.png'
				stationAndCity = trip.DepAirp.split("(")
				depAirp = stationAndCity[0].trim()
				depCity = stationAndCity[1].split("-")[0].trim()
				stationAndCity = trip.ArrAirp.split("(")
				arrAirp = stationAndCity[0].trim()
				arrCity = stationAndCity[1].split("-")[0].trim()
			}

			if (trip.FlightNumber.substring(0,2)=='LA'){ //lan
				flightNumberAirline = trip.FlightNumber + ' <small>(LATAM)</small>'
			} else if (trip.FlightNumber.substring(0,2)=='AR') { //aerolineas
				flightNumberAirline = trip.FlightNumber + ' <small>(Aerolíneas Arg.)</small>'
			} else { // bus
				var splited = trip.FlightNumber.split("-")
				flightNumberAirline = splited.slice(0,splited.length-1).join('')  + '<small>(' + splited[splited.length-1].trim() + ')</small>'
			}

			rowStyle = ''
			if (j%2 == 0) {
				// change row color
				rowStyle = ' style="background-color:#F0F0F0"'
			}

			modal += '<div class="row"' + rowStyle + '>\
						<div class="span12">\
                            <div class="trip-dep-arr-container container">\
                                <div class="row title-row">\
                                    <img src="static/assets/'+pngTransp+'"><div class="line"></div><p>'+flightNumberAirline+'</p>\
                                </div>\
                                <div class="row">\
                                    <div class="col-md-12" style="text-align: center">\
                                        <p style="margin:2px">'+depTime+'   -   '+arrTime+plusOneDay+'</p>\
                                    </div>\
                                </div>\
                                <div class="row">\
                                    <div class="col-md-3">\
                                        <div class="hour-dep-arr"><strong>'+depCity+'</strong></br><small>('+depAirp+')</small></div>\
                                    </div>\
                                    <div class="col-md-6 line"></div>\
                                    <div class="col-md-3">\
                                        <div class="hour-dep-arr"><strong>'+arrCity+'</strong></br><small>('+arrAirp+')</small></div>\
                                    </div>\
                                </div>\
                            </div>\
                            <div class="trip-price-container">\
                                <div class="price">'+addPoints(price)+'$</br></div>\
                                <div class="price-per-adult">('+addPoints(priceAdult)+'$ p/adulto)</div>\
                                <form action="'+trip.Url+'" method="POST" target="_blank">'
                                	if (trip.UrlParams != null) {
                                		for (var k in trip.UrlParams) {
                                			modal += '<input type="hidden" name="'+k+'" value="'+trip.UrlParams[k]+'">'
                                		}
                                	}
                                    modal += '<button type="submit" class="btn btn-info btn-price">Comprar</button>\
                                </form>\
                            </div>\
                        </div>\
                    </div>'
		}

		modal += '</div>\
                      <div class="modal-footer">\
                        <button type="button" class="btn btn-default" data-dismiss="modal">Cerrar</button>\
                      </div>\
                    </div>\
                  </div>\
                </div>\
            </div>\
        </div>\
    </div>'

		well = '<div class="span12 well trip" id="'+tripId+'">\
	                <div class="trip-dep-arr-container container">\
	                    <div class="row">\
	                        <div class="col-md-6 col-dep-arr col-md-offset-3">\
	                            <div class="dep-arr-content">\
	                                <div class="transportation-dep-arr text-center">'
	                                for (k = 0; k < transpTypes.length; k++) {
	                                	type = transpTypes[k]
	                                	if (k != 0) {
	                                		well += '<div class="transp-img transp-img-sum"><img class="img-responsive" src="static/assets/plus-symbol.png"></div>'
	                                	}
	                                	if (type == 0) {
	                                		well += '<div class="transp-img transp-img-transp"><img class="img-responsive" src="static/assets/airplane-flight.png"></div>'
	                                	} else if (type == 1) {
	                                		well += '<div class="transp-img transp-img-transp"><img class="img-responsive" src="static/assets/bus.png"></div>'
	                                	}
	                                }
	                                well += '</div>\
	                            </div>\
	                        </div>\
	                        <div class="col-md-3"></div>\
	                    </div>\
	                    <div class="row">\
	                        <div class="col-md-3">\
	                            <div class="hour-dep-arr"><strong>'+("0" + depHour).slice(-2)+':'+("0" + depMin).slice(-2)+'</strong></br><small>'+("0" + depDay).slice(-2)+'/'+("0" + depMonth).slice(-2)+'</small></div>\
	                        </div>\
	                        <div class="col-md-6 line"></div>\
	                        <div class="col-md-3">\
	                            <div class="hour-dep-arr"><strong>'+("0" + arrHour).slice(-2)+':'+("0" + arrMin).slice(-2)+'</strong></br><small>'+("0" + arrDay).slice(-2)+'/'+("0" + arrMonth).slice(-2)+'</small></div>\
	                        </div>\
	                    </div>\
	                </div>\
	                <div class="trip-price-container">\
	                    <div class="trip-price text-center">\
	                        <div class="price">'+addPoints(totalPrice)+'$</br></div>\
	                        <div class="price-per-adult">('+addPoints(pricePerAdult)+'$ p/adulto)</div>\
	                        </div><div class="btn-price-container"><button type="button" class="btn btn-success" data-toggle="modal" data-target="#myModal'+tripId+'">Ver</button>'

         newTrips = newTrips + well + modal
	}

	allHtmlTrips = []
	var htmlNewTrips = $(newTrips)
	for (var i = 0; i < htmlNewTrips.length; i++) {
		allHtmlTrips.push(htmlNewTrips[i])
	}
	for (var i = 0; i < htmlCurrentTrips.length; i++) {
		allHtmlTrips.push(htmlCurrentTrips[i])
	}

	// sort the trips
	var allHtmlTripsObj = {html: allHtmlTrips}
	sortTrips(allHtmlTripsObj)
	allHtmlTrips = allHtmlTripsObj.html
	
	// insert, and get info for the filters
	var maxPrice = 0
	var minPrice = 100000000
	var maxDepTime = 0
	var minDepTime = 100000000
	var maxTransf = 0
	for (var i = 0; i < allHtmlTrips.length; i++) {
        document.getElementById('trips').appendChild(allHtmlTrips[i]);

        // price filter
        var totalPrice = getTripTotalPrice(allHtmlTrips[i])
        if (totalPrice > maxPrice) {
        	maxPrice = totalPrice
        }
        if (totalPrice < minPrice) {
        	minPrice = totalPrice
        }

        // dep time filter
     	var depHour = getTripDepHourMin(allHtmlTrips[i])
     	var depTime = depHour[0]*60 + depHour[1]
     	if (depTime < minDepTime) {
     		minDepTime = depTime
     	}
     	if (depTime > maxDepTime) {
     		maxDepTime = depTime
     	}

     	// arr day filter
     	var diffDays = getTravelDays(allHtmlTrips[i])
     	if (arrDayFilterDays.indexOf(diffDays) == -1) {
     		arrDayFilterDays.push(diffDays)
     	}

     	// transfers filter
     	var transfers = getTripTransfers(allHtmlTrips[i])
     	if (transfers > maxTransf) {
     		maxTransf = transfers
     	}

    }

    updatePriceFilter(minPrice, maxPrice)
    updateDepHourFilter(minDepTime, maxDepTime)
    updateArrDayFilter()
    updateTransfersFilter(maxTransf)

    updatePagination()

   	$(".main-container-right").removeClass("waitingForTripsLoading")
   	$('.filtering-options').removeClass("waitingForTripsLoading")
   	$('.filter-option').css('display', 'block')

}

function removeDuplicates(oldJson, newJson) {

	var finalJson = []
	for(var i=0; i<newJson.length; i++) {
		var found = false
        for(var j=0; j<oldJson.length; j++) 
            if(compareJSONs(newJson[i], oldJson[j])) found = true
        if (!found) finalJson.push(newJson[i])
    }

    return finalJson

}

// a, b = []Json
function compareJSONs(a, b) {

	if (a.length != b.length) return false

	for (var i = 0; i < a.length; i++) {

		var aJson = a[i]; var bJson = b[i]

		for (var key in aJson) {
			if (key == 'Url' || key == 'UrlParams' || key == 'Id') continue
			if (aJson[key] != bJson[key])
				return false
		}
	}

	return true
}

// increases the progress bar by the given percentage (0.x), and at the end checks whether there are trips or not.
function increaseProgressBar(ind) {

	var length = document.getElementById('trips').children.length

	// increase search counter
	if (ind <= 2) {
		var next = ind+1
		$('#running-search').html("<em>Buscando enlaces directos... (" + next + "/3)</em>") 
	} else if (ind == 3) {
		$('#running-search').html("<em>Buscando combinaciones...</em>") 
	} else {
		$('#running-search em').velocity({opacity:0},400);
	}

	// increase "results"
	document.getElementById('number-of-results').innerHTML = length + ' resultados'

	var parentWidth = $($('.sorting-options-progress').parent()).width()
	var finalWidth = parentWidth * 0.25 * ind

	$('.sorting-options-progress').velocity({width:finalWidth},1000);

	if (ind == 4) {
		$('.sorting-options-progress').fadeOut()

		// check if there are trips in the div. If there aren't, print a message.
		if (length == 0) {
			$(".main-container-right").removeClass("waitingForTripsLoading")
			$('.filtering-options').removeClass("waitingForTripsLoading")
			document.getElementById('trips').innerHTML = '<div class="span12 well"><p class="text-warning" style="text-align: center">Lo sentimos, no se han encontrado combinaciones para este viaje!</br> \
																																		Haz <a href="/index" class="text-warning"><strong>click</strong></a> para volver al inicio</p></div>'
		}

	}

}

function sendPosts() {

	queryInd++

	// restart json
	allTripsJson = []

	// comb json
	firstCombJsonArrived = false
	firstCombJson = null

	// restart "results"
	$('#number-of-results').html('0 resultados')

	// hide filters and erase trips
	$('.filter-option').css('display','none')
	$('#trips').html('')

	// progress bar
	$('.sorting-options-progress').css('opacity', '1.0')
	$('.sorting-options-progress').css('width', '0%')
	$('.sorting-options-progress').css('display', 'block')

	// current page number
	pageNumber = 1
	$('.pages').bootpag({total: pageNumber});

	depYear = 0

	// filter sliders have changed
	priceFilterChanged = false
	depHourFilterChanged = false
	arrDayFilterDays = [-1]
	transfersFilterChanged = false

	depYear = year

    $('.dep-city').html('<strong>'+depName+'</strong>')
    $('.arr-city').html('<strong>'+arrName+'</strong>')
    var ppl = parseInt(adults)+parseInt(children5)+parseInt(children11)+parseInt(infants)
    var persona = 'persona'
    if (ppl > 1) persona = 'personas'
    $('.number-people').html('<strong>'+ppl+'</strong> x <img src="static/assets/man-silhouette.png">')
    $('.dep-date-info').html(("0" + day).slice(-2) + '/' + ("0" + month).slice(-2) + '/' + year)

    var ind = 0

    // set loading gif
    $trips = $(".main-container-right")
    $trips.addClass("waitingForTripsLoading")
    $('.filtering-options').addClass("waitingForTripsLoading")

    var thisQueryInd = queryInd


    // get direct trips
    $.ajax( {
        type: "POST",
        url: "/directTripsPlat10",
        data: JSON.stringify({
            Year: year,
            Month: month,
            Day: day,
            DepId: dep,
            ArrId: arr,
            Adults: adults,
            Children11: children11,
            Children5: children5,
            Infants: infants
        }),
        contentType: "application/json",
        success: function (result) {
        	if (thisQueryInd == queryInd) { // != search changed
        		ind++
	            addCombinationsSorted(JSON.parse(result), depName, arrName);
	            increaseProgressBar(ind)
        	}
        },
        error: function (xhr, ajaxOptions, thrownError) {
            if (thisQueryInd == queryInd) { // != search changed
            	ind++
	            increaseProgressBar(ind)
	            console.log(xhr.status);
	            console.log(thrownError);
	        }
        }
    });

    $.ajax( {
        type: "POST",
        url: "/directTripsAerolineas",
        data: JSON.stringify({
            Year: year,
            Month: month,
            Day: day,
            DepId: dep,
            ArrId: arr,
            Adults: adults,
            Children11: children11,
            Children5: children5,
            Infants: infants
        }),
        contentType: "application/json",
        success: function (result) {
        	if (thisQueryInd == queryInd) { // != search changed
	            ind++
	            addCombinationsSorted(JSON.parse(result), depName, arrName);
	            increaseProgressBar(ind)
	        }
        },
        error: function (xhr, ajaxOptions, thrownError) {
            if (thisQueryInd == queryInd) { // != search changed
            	ind++
	            increaseProgressBar(ind)
	            console.log(xhr.status);
	            console.log(thrownError);
	        }
        }
    });

    $.ajax( {
        type: "POST",
        url: "/directTripsLAN",
        data: JSON.stringify({
            Year: year,
            Month: month,
            Day: day,
            DepId: dep,
            ArrId: arr,
            Adults: adults,
            Children11: children11,
            Children5: children5,
            Infants: infants
        }),
        contentType: "application/json",
        success: function (result) {
        	if (thisQueryInd == queryInd) { // != search changed
	            ind++
	            addCombinationsSorted(JSON.parse(result), depName, arrName);
	            increaseProgressBar(ind)
	        }
        },
        error: function (xhr, ajaxOptions, thrownError) {
            if (thisQueryInd == queryInd) { // != search changed
	            ind++
	            increaseProgressBar(ind)
	            console.log(xhr.status);
	            console.log(thrownError);
	        }
        }
    });


    // usual combinations + specific day combinations

    var usualTrips = ''
    var specificTrips = ''

    $.ajax( {
        type: "POST",
        url: "/usualCombinations",
        data: JSON.stringify({
            Year: year,
            Month: month,
            Day: day,
            DepId: dep,
            ArrId: arr,
            Adults: adults,
            Children11: children11,
            Children5: children5,
            Infants: infants
        }),
        contentType: "application/json",
        success: function (result1) {
            // specific day
            // usualTrips = result1
            if (thisQueryInd != queryInd) { // != search changed
            	return
			}
            if (!firstCombJsonArrived) {
 	           	firstCombJsonArrived = true
 	           	if (result1 != null) {
 	           		firstCombJson = JSON.parse(result1)
 	           		addCombinationsSorted(firstCombJson, depName, arrName)
 	           	}
            } else {
            	ind++
            	if (result1 != null) {
            		var jsonFinal = []
	            	if (firstCombJson != null) {
		            	var newJson = JSON.parse(result1)
		            	jsonFinal = removeDuplicates(firstCombJson, newJson)
		            } else {
		            	jsonFinal = JSON.parse(result1)
		            }
	            	addCombinationsSorted(jsonFinal, depName, arrName)
	            }
	            increaseProgressBar(ind)
            }
        },
        error: function (xhr, ajaxOptions, thrownError) {
            if (thisQueryInd == queryInd) { // != search changed
            	ind++
                increaseProgressBar(ind)
                console.log(xhr.status);
                console.log(thrownError);
            }
        }
    }); 
    $.ajax( {
        type: "POST",
        url: "/sameDayCombinations",
        data: JSON.stringify({
            Year: year,
            Month: month,
            Day: day,
            DepId: dep,
            ArrId: arr,
            Adults: adults,
            Children11: children11,
            Children5: children5,
            Infants: infants
        }),
        contentType: "application/json",
        success: function (result2) {
            // specificTrips = result2
            // removeRepeated(result1, result2)
            if (thisQueryInd != queryInd) { // != search changed
            	return
			}
            if (!firstCombJsonArrived) {
 	           	firstCombJsonArrived = true
 	           	if (result2 != null) {
 	           		firstCombJson = JSON.parse(result2)
 	           		addCombinationsSorted(firstCombJson, depName, arrName)
 	           	}
            } else {
            	ind++
            	if (result2 != null) {
            		var jsonFinal = []
	            	if (firstCombJson != null) {
	            		var newJson = JSON.parse(result2)
		            	jsonFinal = removeDuplicates(firstCombJson, newJson)
	            	} else {
	            		jsonFinal = JSON.parse(result2)
	            	}
	            	addCombinationsSorted(jsonFinal, depName, arrName)
	            }
	            increaseProgressBar(ind)
            }
        },
        error: function (xhr, ajaxOptions, thrownError) {
            if (thisQueryInd == queryInd) { // != search changed
            	ind++
                increaseProgressBar(ind)
                console.log(xhr.status);
                console.log(thrownError);
            }
        }
    }); 

   //  $.when( ajax1(), ajax2() ).then(
   //      function() {
   //          // ajax1 AND ajax2 succeeded

   //          if (thisQueryInd != queryInd) { // != search changed
   //          	return
			// }
   //          ind++

   //          jsonUsual = JSON.parse(usualTrips)
   //          jsonSpecific = JSON.parse(specificTrips)

   //          if (jsonUsual == null && jsonSpecific == null) {
   //              increaseProgressBar(ind)
   //              return
   //          }

   //          var jsonFinal

   //          if (jsonUsual == null) {
   //              jsonFinal = jsonSpecific
   //          } else if (jsonSpecific == null) {
   //              jsonFinal = jsonUsual
   //          } else {
   //              jsonTotal = jsonUsual.concat(jsonSpecific)
   //              jsonFinal = removeDuplicates(jsonTotal)
   //          }

   //          addCombinationsSorted(jsonFinal, depName, arrName)

   //          increaseProgressBar(ind)

   //      },
   //      function() {
   //          // ajax1 OR ajax2 succeeded
   //      }
   //  );
}

function changeSearch() {

	var idOrig = document.getElementById('select-orig-city').value;
	var idDest = document.getElementById('select-dest-city').value;

	alerts = false

	if (idOrig == idDest) {
		 alerts = true
		 $(document.getElementById('select-orig-city').parentNode.children[1].children[0]).css('background-color','#FBF6DC')
		 $(document.getElementById('select-dest-city').parentNode.children[1].children[0]).css('background-color','#FBF6DC')
	} else if (idOrig == '' || idDest == '' ) {
		alerts = true
		if (idOrig == '') { 
			$(document.getElementById('select-orig-city').parentNode.children[1].children[0]).css('background-color','#FBF6DC') 
			$(document.getElementById('select-dest-city').parentNode.children[1].children[0]).css('background-color','white')
		}
		if (idDest == '') { 
			$(document.getElementById('select-dest-city').parentNode.children[1].children[0]).css('background-color','#FBF6DC') 
			$(document.getElementById('select-orig-city').parentNode.children[1].children[0]).css('background-color','white')
		}
	} else {
		$(document.getElementById('select-orig-city').parentNode.children[1].children[0]).css('background-color','white')
		$(document.getElementById('select-dest-city').parentNode.children[1].children[0]).css('background-color','white')
	}

	var fecha = document.getElementById('dep-date').value;
	
	if (fecha == '') {
		alerts = true
		$('#dep-date').css('background-color','#FBF6DC')
	} else {
		$('#dep-date').css('background-color','white')
	}

	if (!alerts) {
		getFormInfo()
		hideChangeSearch()
		sendPosts()
	}
}

function nextDay() {

	var nextDayDate = new Date(parseInt(year), parseInt(month)-1, parseInt(day)+1)
	var nextDay = nextDayDate.getDate()
	var nextDayMonth = nextDayDate.getMonth()+1
	var nextDayYear = nextDayDate.getFullYear()
	year = ''+nextDayYear; month = ''+nextDayMonth; day = ''+nextDay
	sendPosts()
}

function prevDay() {

	var nextDayDate = new Date(parseInt(year), parseInt(month)-1, parseInt(day)-1)
	var nextDay = nextDayDate.getDate()
	var nextDayMonth = nextDayDate.getMonth()+1
	var nextDayYear = nextDayDate.getFullYear()
	year = ''+nextDayYear; month = ''+nextDayMonth; day = ''+nextDay
	sendPosts()
}

function showChangeSearch() {
	$('#change-search').css('display','none')
	$('.change-search-class').css('display','flex')
  	$('#change-search-close').css('display', 'block')
}

function hideChangeSearch() {
	$('#change-search').css('display','block')
	$('.change-search-class').css('display','none')
  	$('#change-search-close').css('display', 'none')
}

function setDepartureDate() {
	var today = new Date();
    var dd = today.getDate();
    var mm = today.getMonth()+1; //January is 0!
    var yyyy = today.getFullYear();

    if(dd<10) {
        dd='0'+dd
    } 

    if(mm<10) {
        mm='0'+mm
    } 

    today = dd+'/'+mm+'/'+yyyy;
    document.getElementById('dep-date').value = today
}