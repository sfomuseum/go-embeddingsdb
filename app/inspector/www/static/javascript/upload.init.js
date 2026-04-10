window.addEventListener('load', function(e){

    const main = document.querySelector("#main");
    const record_uri = main.getAttribute("data-record-uri");    
    const api_upload_uri = main.getAttribute("data-api-upload-uri");
    
    const target = document.querySelector("#upload-similar");
    const upload_image = document.querySelector("#upload-image");
    const upload_input = document.querySelector("#upload");
    const model_input = document.querySelector("#upload-model-select");    
    
    const progress = document.querySelector("#progress");
    const spinner = document.querySelector("#upload-spinner-svg");        
    const submit = document.querySelector("#submit");

    upload_input.addEventListener('change', function(){
	
	const file = this.files[0];

	if (!file) {
	    return;
	}

	const reader = new FileReader();

	reader.onload = function (e) {

	    const im = document.createElement("img");
	    im.src = e.target.result;

	    upload_image.innerHTML = "";
	    upload_image.appendChild(im);
	    upload_image.style.display = "block";
	    
	    // Result any existing similar results
	    target.innerHTML = "";	    
	};
	
	reader.readAsDataURL(file);
    });
    
    const display_similar = function(data){

	// TBD: Replace as much of this code as possible with HTML <template> elements?
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/template
	
	const count = data.length;
	
	const table_headers = [
	    "Distance",
	    "Provider",
	    "Depiction",
	    "Subject",
	    "Attributes",
	];
	
	const count_headers = table_headers.length;
	
	const grid_el = document.createElement("div");
	
	for (var i=0; i < count; i++){
	    
	    const similar = data[i];
	    
	    const image_el = document.createElement("div");
	    image_el.setAttribute("class", "similar-image");
	    
	    if (similar.attributes["preview"]){
		
		const img = document.createElement("img");
		img.setAttribute("src", similar.attributes["preview"]);
		image_el.appendChild(img);
	    }
	    
	    grid_el.appendChild(image_el);

	    // START OF properies (header)
	    
	    const table_el = document.createElement("table");
	    table_el.setAttribute("class", "table similar-table");

	    const table_data = {
		"Distance": similar.similarity,
		"Provider": similar.provider,
		"Depiction": similar.depiction_id,
		"Subject": similar.subject_id,				
	    };

	    for (const k in table_data) {
		
		const v = table_data[k];
		
		const row = document.createElement("tr");
		
		const header = document.createElement("th");
		header.appendChild(document.createTextNode(k));
		row.appendChild(header);
		
		const cell = document.createElement("td");

		if (k == "Depiction"){

		    const depiction_url = new URL("/", location);
		    const depiction_params = new URLSearchParams();
		    
		    depiction_url.pathname = record_uri + similar.provider + "/" + similar.depiction_id;
		    depiction_params.set("model", model_input.value);
		    
		    depiction_url.search = depiction_params;
		    
		    const depiction_link = document.createElement("a");
		    depiction_link.setAttribute("href", depiction_url.toString());
		    depiction_link.appendChild(document.createTextNode(similar.depiction_id));

		    cell.appendChild(depiction_link);
		    
		} else {
		    
		    cell.appendChild(document.createTextNode(v));
		}

		row.appendChild(cell);
		table_el.appendChild(row);
	    }
		    
	    // START OF attributes

	    const attrs_row = document.createElement("tr");
	    const attrs_header = document.createElement("th");
	    attrs_header.appendChild(document.createTextNode("Attributes"));
	    attrs_row.appendChild(attrs_header);
	    
	    const attrs_cell = document.createElement("td");
	    
	    const attrs_table = document.createElement("table");
	    attrs_table.setAttribute("class", "table attributes");
	    
	    for (k in similar.attributes){
		
		const attr_row = document.createElement("tr");
		
		const attr_header = document.createElement("th");
		attr_header.appendChild(document.createTextNode(k));
		attr_row.appendChild(attr_header);
		
		const attr_cell = document.createElement("td");
		attr_cell.appendChild(document.createTextNode(similar.attributes[k]));
		attr_row.appendChild(attr_cell);
		
		attrs_table.appendChild(attr_row);
	    }
	    
	    attrs_cell.appendChild(attrs_table);
	    attrs_row.appendChild(attrs_cell);
	    table_el.appendChild(attrs_row);
	    
	    // END OF attributes

	    // END OF properties (header and data)
	    
	    grid_el.appendChild(table_el);		
	}
	
	target.innerHTML = "";

	var summary;

	switch (count){
	    case 0:
		summary = "There are no matching records.";
		break;
	    case 1:
		summary = "There is one similar record.";
		break;
	    default:
		summary = "There are " + count + " similar records.";
		break;
	}

	const summary_el = document.createElement("div");
	summary_el.setAttribute("id", "summary");
	summary_el.appendChild(document.createTextNode(summary));

	target.appendChild(summary_el);
	target.appendChild(grid_el);
    };
    

    submit.onclick = function(e){
	
	const u = new URL(api_upload_uri, location);
	const url = u.toString();
	
	const form = document.querySelector("#upload-form");
	const data = new FormData(form);

	const xhr = new XMLHttpRequest();
	xhr.open('POST', url);

	xhr.upload.addEventListener('progress', function (ev) {
	    
	    if (ev.lengthComputable) {
		const percent = Math.round((ev.loaded / ev.total) * 100);
		progress.value = percent;
		console.debug(`Upload progress: ${percent}%`);

		if (percent == 100){
		    progress.style.display = "none";
		    progress.value = 0;
		}
	    }
	});

	xhr.addEventListener('load', function () {

	    spinner.style.display = "none";	    
	    progress.style.display = 'none';
	    progress.value = 0;
	    
	    if (xhr.status != 200){

		const feedback_el = document.createElement("div");
		feedback_el.setAttribute("class", "error");
		feedback_el.appendChild(document.createTextNode("There was a problem processing your upload: " + xhr.statusText));
		target.appendChild(feedback_el);
		
		console.error('Upload failed', xhr.status, xhr.statusText);
		return;
	    }

	    var data;
	    
	    try {
		data = JSON.parse(xhr.responseText);
	    } catch (err) {
		console.error("Failed to parse response", err);
		return;
	    }

	    display_similar(data);
	});

	target.innerHTML = "";
	// progress.style.display = "inline-block";
	spinner.style.display = "inline-block";
	
	xhr.send(data);
	return false;
    };

    submit.removeAttribute("disabled");
});
