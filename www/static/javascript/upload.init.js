window.addEventListener('load', function(e){

    const upload = document.querySelector("#upload");
    const file = document.querySelector("#model");
    const similar_provider = document.querySelector("#similar-provider");
    const submit = document.querySelector("#submit");

    submit.onclick = function(e){

	const url = new URL("/upload", location);
	
	const form = e.target;
	const data = new FormData(form);

	const xhr = new XMLHttpRequest();
	xhr.open('POST', u.toString());

	xhr.upload.addEventListener('progress', function (ev) {
	    
	    if (ev.lengthComputable) {
		const percent = Math.round((ev.loaded / ev.total) * 100);
		// progressBar.value = percent;
		console.log(`Upload progress: ${percent}%`);
	    }
	});

	xhr.addEventListener('load', function () {
	    progressBar.style.display = 'none';
	    if (xhr.status >= 200 && xhr.status < 300) {
		console.log('Upload succeeded', xhr.responseText);
	    } else {
		console.error('Upload failed', xhr.status, xhr.statusText);
	    }
	});

	xhr.send(form);
	console.log("Upload");
	return false;
    };

    submit.removeAttribute("disabled");
});
