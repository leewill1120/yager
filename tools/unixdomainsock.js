http=require('http')
req = http.request(
	{
		protocol: 'http:',
		method:'POST',
		path:'/Plugin.Activate',
		socketPath: '/run/docker/plugins/yager.sock'
	}, function(res){
	res.on('data', function(chunk){
		console.log(`${chunk}`)
	});
})
req.end()