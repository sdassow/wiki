
function appendUpload(data) {
	let ul = document.querySelector('#file-list');

	if ( ! ul ) {
		return;
	}

	for ( let n = 0; n < data.length; n++ ) {
		let fileName = data[n].Info.Name;
		let filePath = data[n].Dir;

		let li = document.createElement('li');

		let a = document.createElement('a');
		a.appendChild(document.createTextNode(fileName));
		a.href = '/file/' + filePath + '/' + fileName;

		li.appendChild(a);
		li.appendChild(document.createTextNode(' (' + data[n].Info.Size + ')'))
		ul.appendChild(li);
	}
}

function startUpload(endpoint) {
	let progress = document.querySelector('#file-progress');
	let input = document.querySelector('#file-input');
	if (input.files.length == 0) {
		progress.value = 0;
		return;
	}
	doUpload(endpoint, input.files);
}

function doUpload(endpoint, files, onSuccess, onError) {
	let data = new FormData();
	for ( let n = 0; n < files.length; n++ ) {
		data.append( 'file', files[n] );
	}

	let progress = document.querySelector('#file-progress');
	let req = new XMLHttpRequest();
	req.open('POST', endpoint);
	req.upload.addEventListener('progress', (e) => {
		let completed = (e.loaded / e.total) * 100;
		if (progress) {
			progress.value = completed;
		}
	})
	req.addEventListener('load', (e) => {
		console.log('status', req.status);
		console.log('res', req.response);

		if (req.status == 200) {
			let data = JSON.parse(req.response);
			appendUpload(data);
			if (onSuccess) {
				let fileName = data[0].Info.Name;
				let filePath = data[0].Dir;
				//let imgurl = '/file/' + filePath + '/' + fileName;
				let imgurl = '/' + filePath + '/' + fileName;
				onSuccess(imgurl);
			}
		}
	})
	req.send(data);
}

