URL := http://nevers.studlab.os3.nl:8080
OUTFILE := out.jpg

IN := --data-binary "@chrome2x.jpg"
OUT := -o ${OUTFILE}
HEADER :=  --header "Content-Type: image/jpeg" --header "Accept: image/jpeg"

CURLOPTS := -L -v
CURL := curl ${CURLOPTS} ${IN} ${HEADER} ${OUT}

clean:
	-@rm ${OUTFILE} 2>/dev/null || true
test:
	${CURL} ${URL}

# Working
zeit-now-warm: clean
	${CURL} https://warm.lsproject.now.sh/api/zeit
zeit-now-cold: clean
	${CURL} https://cold.lsproject.now.sh/api/zeit

gcp-run-warm: clean
	${CURL} https://warm-6jdjoh342a-ew.a.run.app/
gcp-run-cold: clean
	${CURL} https://cold-6jdjoh342a-ew.a.run.app/

gcp-fun-warm: clean
	${CURL} https://europe-west1-cedar-channel-259712.cloudfunctions.net/warm
gcp-fun-cold: clean
	${CURL} https://europe-west1-cedar-channel-259712.cloudfunctions.net/cold

azure-fun-warm: clean
	${CURL} https://lsproject.azurewebsites.net/api/warm
azure-fun-cold: clean
	${CURL} https://lsproject.azurewebsites.net/api/cold

ibm-fun-warm: clean
	${CURL} https://eu-gb.functions.cloud.ibm.com/api/v1/web/marbadias97%40gmail.com_dev/default/warm
ibm-fun-cold: clean
	${CURL} https://eu-gb.functions.cloud.ibm.com/api/v1/web/marbadias97%40gmail.com_dev/default/cold

ali-fun-warm: clean
	${CURL} https://5055975195697149.eu-central-1.fc.aliyuncs.com/2016-08-15/proxy/warm/warm/
ali-fun-cold: clean
	${CURL} https://5055975195697149.eu-central-1.fc.aliyuncs.com/2016-08-15/proxy/cold/cold/

gcp-app-warm: clean
	${CURL} https://warm-dot-cedar-channel-259712.appspot.com/
gcp-app-cold: clean
	${CURL} https://cold-dot-cedar-channel-259712.appspot.com/

# broken
aws-lambda-warm: clean
	${CURL} https://xt8gn4dt7g.execute-api.eu-west-2.amazonaws.com/default/warm
aws-lambda-cold: clean
	${CURL} https://wo3u7x3tt8.execute-api.eu-west-2.amazonaws.com/default/cold
