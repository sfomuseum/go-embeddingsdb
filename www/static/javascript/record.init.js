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
	const m = el.value;

	if (m == current_model){
	    return false;
	}

	var href= "/record/" + current_provider + "/" + current_depiction_id + "?model=" + m;

	if (model_provider.value != ""){
	    href = href + "&similar-provider=" + model_provider.value;
	}
	
	console.log(href);
	
	location.href= href;
	return false;
    };
    
    model_current.style.display = "none";
    model_select.style.display = "block";    

    //

    model_provider.onchange = function(e){

	const el = e.target;
	const p = el.value;

	if (p == current_provider){
	    return false;
	}

	const m = model_select.value;

	var href = "/record/" + current_provider + "/" + current_depiction_id + "?model=" + m;

	if (p != "") {
	    href = href + "&similar-provider=" + p;
	}
	
	location.href= href
	return false;
    };

    similar_controls.style.display = "block";
});
