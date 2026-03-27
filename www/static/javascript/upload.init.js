window.addEventListener('load', function(e){

    const progress = document.querySelector("#progress");    
    const submit = document.querySelector("#submit");

    submit.onclick = function(e){

	const u = new URL("/api/upload/", location);
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
	    }
	});

	xhr.addEventListener('load', function () {
	    
	    progress.style.display = 'none';
	    progress.value = 0;
	    
	    if (xhr.status != 200){
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
	    
	    const count = data.length;

	    for (var i=0; i < count; i++){

		const similar = data[i];
		console.log("similar", similar);
	    }
	    
	});

	progress.style.display = "block";

	xhr.send(data);
	return false;
    };

    submit.removeAttribute("disabled");
});
