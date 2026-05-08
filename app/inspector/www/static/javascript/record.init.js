window.addEventListener('load', function(e){

    const main = document.querySelector("#main");
    const providers_uri = main.getAttribute("data-api-providers-uri");
    const models_uri = main.getAttribute("data-api-models-uri");
    const record_uri = main.getAttribute("data-record-uri");

    const record_provider = main.getAttribute("data-record-provider");
    
    const model_current = document.querySelector("#model-current");
    const model_select = document.querySelector("#model-select");
    const record_table = document.querySelector("#record-table");
    
    const current_provider = record_table.getAttribute("data-provider");
    const current_depiction_id = record_table.getAttribute("data-depiction-id");
    const current_model = record_table.getAttribute("data-model");        

    const similar_controls = document.querySelector("#similar-controls");
    const provider_select = document.querySelector("#provider-select");

    const max_distance = document.querySelector("#max-distance");
    const max_distance_wrapper = document.querySelector("#max-distance-wrapper");
    const custom_max_distance = document.querySelector("#custom-max-distance");    

    const record_summary = document.querySelector("#record-summary");    
    const record_details = document.querySelector("#record-details");

    record_details.addEventListener("toggle", function(){

	if (record_details.open){
	    record_summary.style.display = "none";
	} else {
	    record_summary.style.display = "block";
	}
	
	return false;
    });
    
    custom_max_distance.addEventListener("change", function(){

	if (custom_max_distance.checked){
	    max_distance_wrapper.style.display = "block";
	} else {
	    max_distance_wrapper.style.display = "none";
	}
	
	return false;	
    });
    
    max_distance.addEventListener("input", function() {
	const el = document.querySelector("#max-distance-value");
	el.textContent = max_distance.value;	
    });
    
    
    const refine_btn = document.querySelector("#refine");

    model_select.onchange = function(){

	const u = new URL(providers_uri, location);
	const s = new URLSearchParams();

	s.set("model", model_select.value);
	u.search = s

	fetch(u.toString())
	    .then(rsp => {
		return rsp.json();
	    }).then(data => {

		provider_select.innerHTML = "";

		const opt = document.createElement("option");
		opt.setAttribute("value", "");
		opt.appendChild(document.createTextNode("All providers"));
		provider_select.appendChild(opt);
		
		const count = data.length;

		for (var i=0; i < count; i++){

		    const opt = document.createElement("option");
		    opt.setAttribute("value", data[i]);
		    opt.appendChild(document.createTextNode(data[i]));
		    provider_select.appendChild(opt);
		}
		
	    }).catch(err => {
		console.error("Failed to get model providers", err)
	    });
	
	return false;
    };

    provider_select.onchange = function(){

	const u = new URL(models_uri, location);
	const s = new URLSearchParams();

	s.set("provider", provider_select.value);
	u.search = s;

	console.debug("Get models for provider", u.toString());
	
	fetch(u.toString())
	    .then(rsp => {
		return rsp.json();
	    }).then(data => {

		// now get models for record provider

		const u2 = new URL(models_uri, location);
		const s2 = new URLSearchParams();
		
		s2.set("provider", record_provider);
		u2.search = s2;

		console.debug("Get record models", u2.toString());
		
		fetch(u2.toString())
		    .then(rsp => {
			return rsp.json();
		    }).then(record_models => {
		
			model_select.innerHTML = "";
			
			const count = data.length;
			
			for (var i=0; i < count; i++){
			    
			    const opt = document.createElement("option");
			    opt.setAttribute("value", data[i]);

			    if (! record_models.includes(data[i])){
				opt.setAttribute("disabled", "disabled");
			    }
			    
			    opt.appendChild(document.createTextNode(data[i]));
			    model_select.appendChild(opt);
			}
			
		    }).catch(err => {
			console.error("Failed to get record models", err);
			throw err;
		    });
		
	    }).catch(err => {
		console.error("Failed to derive provider models", err)
	    });
	
	return false;
    };
    
    refine_btn.onclick = function(){

	const u = new URL("/", location);
	const s = new URLSearchParams();
	
	u.pathname = record_uri + current_provider + "/" + current_depiction_id;

	if (model_select.value != ""){
	    s.set("model", model_select.value)
	}
	
	if (provider_select.value != ""){
	    s.set("similar-provider", provider_select.value);
	}

	if (custom_max_distance.checked){
	    s.set("custom-max-distance", "true");
	    s.set("max-distance", max_distance.value);
	}
	
	u.search = s;
	const href = u.toString();

	console.log("HREF", href);
	location.href= href;
	return false;
	
    }

    const u = new URL(location.href);

    if (u.searchParams.has("custom-max-distance")){
	custom_max_distance.checked = true;
	max_distance_wrapper.style.display = "block";
    }
    
    similar_controls.style.display = "block";
});
