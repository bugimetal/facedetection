function DrawImagePreviews(images, parentID) {
  if (images === undefined) {
    return
  }

  let parent = document.getElementById(parentID)

  images.forEach(image => {
    let img = document.createElement('img');
    img.src = image.url;
    img.onclick = showImage.bind(null, image);

    parent.appendChild(img);
  });
}


function showImage(image, e) {
  e.stopPropagation();

  // making demo elements visible
  document.getElementById('demo').style.visibility = 'visible';

  let imageToDetectPlaceholder = document.getElementById('image-to-detect');
  // remove previous image
  imageToDetectPlaceholder.innerHTML = '';

  let img = document.createElement('img');
  img.src = image.url;
  img.id = 'demo-image';
  imageToDetectPlaceholder.appendChild(img);

  let detectFacesButton = document.getElementById('detect-faces-button');
  detectFacesButton.onclick = FetchAndDrawFaces.bind(null, image);

  let canvas = document.createElement('canvas');
  canvas.id = 'demo-canvas';
  canvas.width = image.width;
  canvas.height = image.height;

  imageToDetectPlaceholder.appendChild(canvas);
}

async function FetchAndDrawFaces(image, e) {
  e.stopPropagation();

  let img = document.getElementById('demo-image');
  let canvas = document.getElementById('demo-canvas');

  canvas.style.position = 'absolute';
  canvas.style.left = img.offsetLeft + 'px';
  canvas.style.top = img.offsetTop + 'px';

  let url = 'http://localhost:8080/v1/facedetection/' + btoa('http://localhost:8080' + image.url);

  let response = await fetch(url);
  let detected = await response.json();

  // handling errors
  if (detected.faces === undefined) {
    let detectError = document.getElementById('detect-error');

    if (detected.error === undefined) {
      detectError.innerHTML = 'unexpected error';
      return
    }

    detectError.innerHTML = detected.error.message;
    return
  }

  let ctx = canvas.getContext('2d');
  ctx.lineWidth = 3;
  ctx.strokeStyle = '#00ff00';
  ctx.beginPath();

  detected.faces.forEach(face => {
    let bounds = face.bounds;
    let rightEye = face.right_eye;
    let leftEye = face.left_eye;
    let mouth = face.mouth;

    // draw face rectangle
    ctx.rect(bounds.x, bounds.y, bounds.width, bounds.height);

    // draw right eye
    ctx.moveTo(rightEye.x + rightEye.scale, rightEye.y);
    ctx.arc(rightEye.x, rightEye.y, rightEye.scale, 0, 2*Math.PI, true);

    // draw left eye
    ctx.moveTo(leftEye.x  + leftEye.scale, leftEye.y);
    ctx.arc(leftEye.x, leftEye.y, leftEye.scale, 0, 2*Math.PI, true);

    // draw mouth
    ctx.rect(mouth.x, mouth.y, mouth.width, mouth.height);

    ctx.stroke();
  });
}