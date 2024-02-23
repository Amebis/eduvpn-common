// Code generated by running "go generate" in golang.org/x/text. DO NOT EDIT.

package client

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

type dictionary struct {
	index []uint32
	data  string
}

func (d *dictionary) Lookup(key string) (data string, ok bool) {
	p, ok := messageKeyToIndex[key]
	if !ok {
		return "", false
	}
	start, end := d.index[p], d.index[p+1]
	if start == end {
		return "", false
	}
	return d.data[start:end], true
}

func init() {
	dict := map[string]catalog.Dictionary{
		"da": &dictionary{index: daIndex, data: daData},
		"de": &dictionary{index: deIndex, data: deData},
		"en": &dictionary{index: enIndex, data: enData},
		"es": &dictionary{index: esIndex, data: esData},
		"fr": &dictionary{index: frIndex, data: frData},
		"it": &dictionary{index: itIndex, data: itData},
		"nl": &dictionary{index: nlIndex, data: nlData},
		"sl": &dictionary{index: slIndex, data: slData},
		"uk": &dictionary{index: ukIndex, data: ukData},
	}
	fallback := language.MustParse("en")
	cat, err := catalog.NewFromMap(dict, catalog.Fallback(fallback))
	if err != nil {
		panic(err)
	}
	message.DefaultCatalog = cat
}

var messageKeyToIndex = map[string]int{
	"An error occurred after getting the discovery files for the list of organizations": 23,
	"An error occurred after getting the discovery files for the list of servers":       24,
	"An internal error occurred":                                                                                                28,
	"Failed to cleanup the VPN connection":                                                                                      18,
	"Failed to get the secure internet server with id: '%s' for setting a location":                                             19,
	"Failed to set the profile ID: '%s'":                                                                                        14,
	"Failover failed to complete with gateway: '%s' and MTU: '%d'":                                                              22,
	"No OAuth tokens were found when cleaning up the connection":                                                                16,
	"No VPN configuration for server: '%s' could be obtained":                                                                   10,
	"Server identifier: '%s', is not valid when getting a VPN configuration":                                                    6,
	"Server identifier: '%s', is not valid when removing the server":                                                            11,
	"Server: '%s' could not be obtained":                                                                                        9,
	"The VPN proxy exited":                                                                                                      25,
	"The client tried to autoconnect to the VPN server: '%s', but the operation failed to complete":                             8,
	"The client tried to autoconnect to the VPN server: '%s', but you need to authorizate again. Please manually connect again": 7,
	"The current server could not be retrieved":                                                                                 13,
	"The current server could not be retrieved when renewing the session":                                                       20,
	"The current server was not found when cleaning up the connection":                                                          15,
	"The current server was not found when getting the VPN expiration date":                                                     1,
	"The custom server with URL: '%s' could not be added":                                                                       4,
	"The institute access server with URL: '%s' could not be added":                                                             2,
	"The log file with directory: '%s' failed to initialize":                                                                    0,
	"The secure internet server with organisation ID: '%s' could not be added":                                                  3,
	"The server was unable to be retrieved when cleaning up the connection":                                                     17,
	"The server was unable to be retrieved when renewing the session":                                                           21,
	"The server: '%s' could not be removed":                                                                                     12,
	"input: '%s' is not a valid URL":                                                                                            5,
	"timeout reached for URL: '%s' and HTTP method: '%s'":                                                                       26,
	"with cause:": 27,
}

var daIndex = []uint32{ // 30 elements
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000,
} // Size: 144 bytes

const daData string = ""

var deIndex = []uint32{ // 30 elements
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000,
} // Size: 144 bytes

const deData string = ""

var enIndex = []uint32{ // 30 elements
	0x00000000, 0x0000003a, 0x00000080, 0x000000c1,
	0x0000010d, 0x00000144, 0x00000166, 0x000001b0,
	0x0000022d, 0x0000028e, 0x000002b4, 0x000002ef,
	0x00000331, 0x0000035a, 0x00000384, 0x000003aa,
	0x000003eb, 0x00000426, 0x0000046c, 0x00000491,
	0x000004e2, 0x00000526, 0x00000566, 0x000005a9,
	0x000005fb, 0x00000647, 0x0000065c, 0x00000696,
	0x000006a2, 0x000006bd,
} // Size: 144 bytes

const enData string = "" + // Size: 1725 bytes
	"\x02The log file with directory: '%[1]s' failed to initialize\x02The cur" +
	"rent server was not found when getting the VPN expiration date\x02The in" +
	"stitute access server with URL: '%[1]s' could not be added\x02The secure" +
	" internet server with organisation ID: '%[1]s' could not be added\x02The" +
	" custom server with URL: '%[1]s' could not be added\x02input: '%[1]s' is" +
	" not a valid URL\x02Server identifier: '%[1]s', is not valid when gettin" +
	"g a VPN configuration\x02The client tried to autoconnect to the VPN serv" +
	"er: '%[1]s', but you need to authorizate again. Please manually connect " +
	"again\x02The client tried to autoconnect to the VPN server: '%[1]s', but" +
	" the operation failed to complete\x02Server: '%[1]s' could not be obtain" +
	"ed\x02No VPN configuration for server: '%[1]s' could be obtained\x02Serv" +
	"er identifier: '%[1]s', is not valid when removing the server\x02The ser" +
	"ver: '%[1]s' could not be removed\x02The current server could not be ret" +
	"rieved\x02Failed to set the profile ID: '%[1]s'\x02The current server wa" +
	"s not found when cleaning up the connection\x02No OAuth tokens were foun" +
	"d when cleaning up the connection\x02The server was unable to be retriev" +
	"ed when cleaning up the connection\x02Failed to cleanup the VPN connecti" +
	"on\x02Failed to get the secure internet server with id: '%[1]s' for sett" +
	"ing a location\x02The current server could not be retrieved when renewin" +
	"g the session\x02The server was unable to be retrieved when renewing the" +
	" session\x02Failover failed to complete with gateway: '%[1]s' and MTU: '" +
	"%[2]d'\x02An error occurred after getting the discovery files for the li" +
	"st of organizations\x02An error occurred after getting the discovery fil" +
	"es for the list of servers\x02The VPN proxy exited\x02timeout reached fo" +
	"r URL: '%[1]s' and HTTP method: '%[2]s'\x02with cause:\x02An internal er" +
	"ror occurred"

var esIndex = []uint32{ // 30 elements
	0x00000000, 0x0000004a, 0x0000004a, 0x0000004a,
	0x0000004a, 0x0000004a, 0x0000004a, 0x0000004a,
	0x0000004a, 0x0000004a, 0x0000004a, 0x0000004a,
	0x0000004a, 0x0000004a, 0x0000004a, 0x0000004a,
	0x0000004a, 0x0000004a, 0x0000004a, 0x0000004a,
	0x0000004a, 0x0000004a, 0x0000004a, 0x0000004a,
	0x000000ab, 0x00000104, 0x00000104, 0x00000104,
	0x00000104, 0x00000104,
} // Size: 144 bytes

const esData string = "" + // Size: 260 bytes
	"\x02El archivo de registro con el directorio: '%[1]s' no se puede inicia" +
	"lizar\x02Se ha producido un error al obtener los archivos de detección d" +
	"e la lista de las organizaciones\x02Se ha producido un error al obtener " +
	"los archivos de detección de la lista de servidores"

var frIndex = []uint32{ // 30 elements
	0x00000000, 0x0000004f, 0x0000004f, 0x0000004f,
	0x000000aa, 0x000000f1, 0x000000f1, 0x000000f1,
	0x000000f1, 0x000000f1, 0x000000f1, 0x000000f1,
	0x000000f1, 0x000000f1, 0x000000f1, 0x000000f1,
	0x000000f1, 0x000000f1, 0x000000f1, 0x000000f1,
	0x000000f1, 0x000000f1, 0x000000f1, 0x000000f1,
	0x0000015c, 0x000001c2, 0x000001c2, 0x000001c2,
	0x000001c2, 0x000001c2,
} // Size: 144 bytes

const frData string = "" + // Size: 450 bytes
	"\x02Le fichier de registre du répertoire\u202f: '%[1]s' n'a pas pu être " +
	"initialisé\x02Le serveur internet sécurisé avec l'ID d'organisation : '%" +
	"[1]s' n'a pas pu être ajouté\x02Le serveur personnalisé avec l'URL : '%[" +
	"1]s' n'a pas pu être ajouté\x02Une erreur est survenue pendant la récupé" +
	"ration des fichiers de détection de la liste des organisations\x02Une er" +
	"reur est survenue pendant la récupération des fichiers de détection de l" +
	"a liste des serveurs"

var itIndex = []uint32{ // 30 elements
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000,
} // Size: 144 bytes

const itData string = ""

var nlIndex = []uint32{ // 30 elements
	0x00000000, 0x0000003c, 0x0000003c, 0x00000083,
	0x000000d0, 0x00000106, 0x00000106, 0x00000106,
	0x00000106, 0x00000106, 0x00000106, 0x00000106,
	0x00000106, 0x00000106, 0x00000106, 0x00000106,
	0x00000106, 0x00000106, 0x00000106, 0x00000106,
	0x00000106, 0x00000106, 0x00000106, 0x00000157,
	0x0000019f, 0x000001e2, 0x000001e2, 0x000001e2,
	0x000001ef, 0x0000020e,
} // Size: 144 bytes

const nlData string = "" + // Size: 526 bytes
	"\x02Het log bestand met pad: '%[1]s' kan niet aangemaakt worden\x02De in" +
	"stitute access server met URL: '%[1]s' kan niet toegevoegd worden\x02De " +
	"secure internet server met identiteit: '%[1]s' kan niet toegevoegd worde" +
	"n\x02De server met URL: '%[1]s' kan niet toegevoegd worden\x02Het 'failo" +
	"ver' proces kan niet voltooid worden. Gateway: '%[1]s' en MTU: '%[2]d'" +
	"\x02Er is een fout opgetreden met het ophalen van de lijst van organisat" +
	"ies\x02Er is een fout opgetreden met het ophalen van de lijst van server" +
	"s\x02met oorzaak:\x02Een interne fout is opgetreden"

var slIndex = []uint32{ // 30 elements
	0x00000000, 0x0000003c, 0x0000003c, 0x00000086,
	0x000000d4, 0x00000110, 0x00000110, 0x00000110,
	0x00000110, 0x00000110, 0x00000110, 0x00000110,
	0x00000110, 0x00000110, 0x00000110, 0x00000110,
	0x00000110, 0x00000110, 0x00000110, 0x00000110,
	0x00000110, 0x00000110, 0x00000110, 0x00000147,
	0x00000187, 0x000001c7, 0x000001c7, 0x000001c7,
	0x000001d1, 0x000001ef,
} // Size: 144 bytes

const slData string = "" + // Size: 495 bytes
	"\x02Napaka pri vzpostavitvi datoteke dnevnika v imeniku '%[1]s'\x02Strež" +
	"nika z naslovom '%[1]s' za dostop do ustanove ni bilo možno dodati\x02St" +
	"režnika za varni splet organizacije z ID-jem '%[1]s' ni bilo možno dodat" +
	"i\x02Svojega strežnika z naslovom '%[1]s' ni bilo možno dodati\x02Preklo" +
	"p ni uspel s prehodom '%[1]s' in MTU-jem '%[2]d'\x02Pri nalaganju datote" +
	"k kataloga organizacij je prišlo do napake\x02Pri nalaganju datotek kata" +
	"loga strežnikov je prišlo do napake\x02; razlog:\x02Prišlo je do notranj" +
	"e napake"

var ukIndex = []uint32{ // 30 elements
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000,
} // Size: 144 bytes

const ukData string = ""

// Total table size 4752 bytes (4KiB); checksum: 1AA1A07B
