function getCities(callback) {
	
	var xmlHttp = new XMLHttpRequest();
    xmlHttp.onreadystatechange = function() { 
        if (xmlHttp.readyState == 4 && xmlHttp.status == 200)
            callback(xmlHttp.responseText);
    }
    xmlHttp.open("GET", "/getDbCities", true); // true for asynchronous 
    xmlHttp.send(null);

}