window.addEventListener('load', function(e){

    const model_current = document.querySelector("#model-current");
    const model_select = document.querySelector("#model-select");
    const record_table = document.querySelector("#record-table");
    
    if (! model_current){
	log.warn("Missing model-current element");
	return;
    }

    if (! model_select){
	log.warn("Missing model-select element");
	return;
    }

    if (! record_table){
	log.warn("Missing record-table element");
	return;
    }

    const current_provider = record_table.getAttribute("data-provider");
    const current_depiction_id = record_table.getAttribute("data-depiction-id");
    const current_model = record_table.getAttribute("data-model");        

    if (! current_provider){
	log.warn("Missing data-provider");
	return;
    }

    if (! current_depiction_id){
	log.warn("Missing data-depiction-id");
	return;
    }

    if (! current_model){
	log.warn("Missing data-model");
	return;
    }

    const similar_controls = document.querySelector("#similar-controls");

    if (! similar_controls){
	log.warn("Missing similar controls");
	return;
    }

    const model_provider = document.querySelector("#model-provider");

    if (! model_provider){
	log.warn("Missing model provider");
	return;
    }

    //
    
    model_select.onchange = function(e){
	const el = e.target;
	const v = el.value;

	if (v == current_model){
	    return false;
	}

	const u = new URL("/", location);
	const s = new URLSearchParams();
	
	u.pathname = "/record/" + current_provider + "/" + current_depiction_id;
	s.set("model",  v);

	if (model_provider.value != ""){
	    s.set("similar-provider", model_provider.value);
	}

	u.search = s;
	const href = u.toString();
	
	location.href= href;
	return false;
    };
    
    model_current.style.display = "none";
    model_select.style.display = "block";    

    //

    model_provider.onchange = function(e){

	const el = e.target;
	const v = el.value;

	if (v == current_provider){
	    return false;
	}

	const u = new URL("/", location);
	const s = new URLSearchParams();
	
	const m = model_select.value;

	u.pathname = "/record/" + current_provider + "/" + current_depiction_id;

	s.set("model", model_select.value);

	if (v != "") {
	    s.set("similar-provider",v);
	}

	u.search = s
	const href = u.toString();
	
	location.href= href
	return false;
    };

    similar_controls.style.display = "block";
});
