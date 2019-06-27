// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

const html = `
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>Photo Frame</title>
    <style>
    	body {
		background-color: #000;
	}
    	.img-container {
		position: absolute;
		top: 0;
		bottom: 0;
		left: 0;
		right: 0;
		background-size: contain;
		background-repeat: no-repeat;
		background-position: center;
		background-image: url("/image");
		-webkit-animation: fadein 2s; /* Safari, Chrome and Opera > 12.1 */
		-moz-animation: fadein 2s; /* Firefox < 16 */
		-ms-animation: fadein 2s; /* Internet Explorer */
		-o-animation: fadein 2s; /* Opera < 12.1 */
		animation: fadein 2s;
	}

	@keyframes fadein {
	    from { opacity: 0; }
	    to   { opacity: 1; }
	}

	/* Firefox < 16 */
	@-moz-keyframes fadein {
	    from { opacity: 0; }
	    to   { opacity: 1; }
	}

	/* Safari, Chrome and Opera > 12.1 */
	@-webkit-keyframes fadein {
	    from { opacity: 0; }
	    to   { opacity: 1; }
	}

	/* Internet Explorer */
	@-ms-keyframes fadein {
	    from { opacity: 0; }
	    to   { opacity: 1; }
	}

	/* Opera < 12.1 */
	@-o-keyframes fadein {
	    from { opacity: 0; }
	    to   { opacity: 1; }
	}
    </style>
  </head>
  <body>
    <div class="img-container">
    </div>
  </body>
<script type="text/javascript">
	(function() {
		window.setTimeout(function() {
			window.location.reload();
		}, {{.Duration}});
	})();
</script>
</html>
`

const loading = `
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>Photo Frame</title>
    <style>
    	body {
		background-color: #000;
	}
	.loading {
		position: absolute;
		color: #fff;
		text-align: center;
		top: 50%;
		width: 100%;
	}
    	.img-container {
		position: absolute;
		top: 0;
		bottom: 0;
		left: 0;
		right: 0;
		background-size: contain;
		background-repeat: no-repeat;
		background-position: center;
	}
    </style>
  </head>
  <body>
    <div class="img-container">
	    <h1 class="loading">Images are loading ...</h1>
    </div>

  </body>
<script type="text/javascript">
	(function() {
		window.setTimeout(function() {
			window.location.reload();
		}, {{.Duration}});
	})();
</script>
</html>
`
