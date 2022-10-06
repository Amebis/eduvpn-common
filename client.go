package eduvpn

import (
	"errors"
	"fmt"

	"github.com/eduvpn/eduvpn-common/internal/config"
	"github.com/eduvpn/eduvpn-common/internal/discovery"
	"github.com/eduvpn/eduvpn-common/internal/fsm"
	"github.com/eduvpn/eduvpn-common/internal/log"
	"github.com/eduvpn/eduvpn-common/internal/oauth"
	"github.com/eduvpn/eduvpn-common/internal/server"
	"github.com/eduvpn/eduvpn-common/types"
	"github.com/eduvpn/eduvpn-common/internal/util"
)

type (
	// ServerBase is an alias to the internal ServerBase
	// This contains the details for each server
	ServerBase = server.ServerBase
)

// Client is the main struct for the VPN client
type Client struct {
	// The language used for language matching
	Language string `json:"-"` // language should not be saved

	// The chosen server
	Servers server.Servers `json:"servers"`

	// The list of servers and organizations from disco
	Discovery discovery.Discovery `json:"discovery"`

	// The fsm
	FSM fsm.FSM `json:"-"`

	// The logger
	Logger log.FileLogger `json:"-"`

	// The config
	Config config.Config `json:"-"`

	// Whether to enable debugging
	Debug bool `json:"-"`
}

// Register initializes the clientwith the following parameters:
//  - name: the name of the client
//  - directory: the directory where the config files are stored. Absolute or relative
//  - stateCallback: the callback function for the FSM that takes two states (old and new) and the data as an interface
//  - debug: whether or not we want to enable debugging
// It returns an error if initialization failed, for example when discovery cannot be obtained and when there are no servers.
func (client *Client) Register(
	name string,
	directory string,
	language string,
	stateCallback func(FSMStateID, FSMStateID, interface{}),
	debug bool,
) error {
	errorMessage := "failed to register with the GO library"
	if !client.InFSMState(STATE_DEREGISTERED) {
		return &types.WrappedErrorMessage{
			Message: errorMessage,
			Err:     FSMDeregisteredError{}.CustomError(),
		}
	}
	// Initialize the logger
	logLevel := log.LOG_WARNING

	// TODO: Verify language setting?
	client.Language = language

	if debug {
		logLevel = log.LOG_INFO
	}

	loggerErr := client.Logger.Init(logLevel, name, directory)
	if loggerErr != nil {
		return &types.WrappedErrorMessage{Message: errorMessage, Err: loggerErr}
	}

	// Initialize the FSM
	client.FSM = newFSM(name, stateCallback, directory, debug)
	client.Debug = debug

	// Initialize the Config
	client.Config.Init(name, directory)

	// Try to load the previous configuration
	if client.Config.Load(&client) != nil {
		// This error can be safely ignored, as when the config does not load, the struct will not be filled
		client.Logger.Info("Previous configuration not found")
	}

	discoServers, discoServersErr := client.GetDiscoServers()

	_, currentServerErr := client.Servers.GetCurrentServer()
	// Only actually return the error if we have no disco servers and no current server
	if discoServersErr != nil && (discoServers == nil || discoServers.Version == 0) && currentServerErr != nil {
		client.Logger.Error(
			fmt.Sprintf(
				"No configured servers, discovery servers is empty and no servers with error: %s",
				GetErrorTraceback(discoServersErr),
			),
		)
		return &types.WrappedErrorMessage{Message: errorMessage, Err: discoServersErr}
	}
	discoOrgs, discoOrgsErr := client.GetDiscoOrganizations()

	// Only actually return the error if we have no disco organizations and no current server
	if discoOrgsErr != nil && (discoOrgs == nil || discoOrgs.Version == 0) && currentServerErr != nil {
		client.Logger.Error(
			fmt.Sprintf(
				"No configured organizations, discovery organizations empty and no servers with error: %s",
				GetErrorTraceback(discoOrgsErr),
			),
		)
		return &types.WrappedErrorMessage{Message: errorMessage, Err: discoOrgsErr}
	}
	// Go to the No Server state with the saved servers
	client.FSM.GoTransitionWithData(STATE_NO_SERVER, client.Servers, true)
	return nil
}

// Deregister 'deregisters' the client, meaning saving the log file and the config and emptying out the client struct.
func (client *Client) Deregister() {
	// Close the log file
	client.Logger.Close()

	// Save the config
	saveErr := client.Config.Save(&client)
	if saveErr != nil {
		client.Logger.Info(
			fmt.Sprintf(
				"Failed saving configuration, error: %s",
				GetErrorTraceback(saveErr),
			),
		)
	}

	// Empty out the state
	*client = Client{}
}

// goBackInternal uses the public go back but logs an error if it happened.
func (client *Client) goBackInternal() {
	goBackErr := client.GoBack()
	if goBackErr != nil {
		client.Logger.Info(
			fmt.Sprintf(
				"Failed going back, error: %s",
				GetErrorTraceback(goBackErr),
			),
		)
	}
}

// GoBack transitions the FSM back to the previous UI state, for now this is always the NO_SERVER state.
func (client *Client) GoBack() error {
	errorMessage := "failed to go back"
	if client.InFSMState(STATE_DEREGISTERED) {
		client.Logger.Error("Wrong state, cannot go back when deregistered")
		return &types.WrappedErrorMessage{
			Message: errorMessage,
			Err:     FSMDeregisteredError{}.CustomError(),
		}
	}

	// FIXME: Abitrary back transitions don't work because we need the approriate data
	client.FSM.GoTransitionWithData(STATE_NO_SERVER, client.Servers, false)
	return nil
}

// ensureLogin logs the user back in if needed.
// It runs the FSM transitions to ask for user input.
func (client *Client) ensureLogin(chosenServer server.Server) error {
	errorMessage := "failed ensuring login"
	// Relogin with oauth
	// This moves the state to authorized
	if server.NeedsRelogin(chosenServer) {
		url, urlErr := server.GetOAuthURL(chosenServer, client.FSM.Name)

		client.FSM.GoTransitionWithData(STATE_OAUTH_STARTED, url, true)

		if urlErr != nil {
			client.goBackInternal()
			return &types.WrappedErrorMessage{Message: errorMessage, Err: urlErr}
		}

		exchangeErr := server.OAuthExchange(chosenServer)

		if exchangeErr != nil {
			client.goBackInternal()
			return &types.WrappedErrorMessage{Message: errorMessage, Err: exchangeErr}
		}
	}
	// OAuth was valid, ensure we are in the authorized state
	client.FSM.GoTransition(STATE_AUTHORIZED)
	return nil
}

// getConfigAuth gets a config with authorization and authentication.
// It also asks for a profile if no valid profile is found.
func (client *Client) getConfigAuth(
	chosenServer server.Server,
	preferTCP bool,
) (string, string, error) {
	loginErr := client.ensureLogin(chosenServer)
	if loginErr != nil {
		return "", "", loginErr
	}
	client.FSM.GoTransition(STATE_REQUEST_CONFIG)

	validProfile, profileErr := server.HasValidProfile(chosenServer)
	if profileErr != nil {
		return "", "", profileErr
	}

	// No valid profile, ask for one
	if !validProfile {
		askProfileErr := client.askProfile(chosenServer)
		if askProfileErr != nil {
			return "", "", askProfileErr
		}
	}

	// We return the error otherwise we wrap it too much
	return server.GetConfig(chosenServer, preferTCP)
}

// retryConfigAuth retries the getConfigAuth function if the tokens are invalid.
// If OAuth is cancelled, it makes sure that we only forward the error as additional info.
func (client *Client) retryConfigAuth(
	chosenServer server.Server,
	preferTCP bool,
) (string, string, error) {
	errorMessage := "failed authorized config retry"
	config, configType, configErr := client.getConfigAuth(chosenServer, preferTCP)
	if configErr != nil {
		level := types.ERR_OTHER
		var error *oauth.OAuthTokensInvalidError
		var oauthCancelledError *oauth.OAuthCancelledCallbackError

		// Only retry if the error is that the tokens are invalid
		if errors.As(configErr, &error) {
			config, configType, configErr = client.getConfigAuth(
				chosenServer,
				preferTCP,
			)
			if configErr == nil {
				return config, configType, nil
			}
		}
		if errors.As(configErr, &oauthCancelledError) {
			level = types.ERR_INFO
		}
		client.goBackInternal()
		return "", "", &types.WrappedErrorMessage{Level: level, Message: errorMessage, Err: configErr}
	}
	return config, configType, nil
}

// getConfig gets an OpenVPN/WireGuard configuration by contacting the server, moving the FSM towards the DISCONNECTED state and then saving the local configuration file.
func (client *Client) getConfig(
	chosenServer server.Server,
	preferTCP bool,
) (string, string, error) {
	errorMessage := "failed to get a configuration for OpenVPN/Wireguard"
	if client.InFSMState(STATE_DEREGISTERED) {
		return "", "", &types.WrappedErrorMessage{
			Message: errorMessage,
			Err:     FSMDeregisteredError{}.CustomError(),
		}
	}

	config, configType, configErr := client.retryConfigAuth(chosenServer, preferTCP)

	if configErr != nil {
		return "", "", &types.WrappedErrorMessage{Level: GetErrorLevel(configErr), Message: errorMessage, Err: configErr}
	}

	currentServer, currentServerErr := client.Servers.GetCurrentServer()
	if currentServerErr != nil {
		return "", "", &types.WrappedErrorMessage{Message: errorMessage, Err: currentServerErr}
	}

	// Signal the server display info
	client.FSM.GoTransitionWithData(STATE_DISCONNECTED, currentServer, false)

	// Save the config
	saveErr := client.Config.Save(&client)
	if saveErr != nil {
		client.Logger.Info(
			fmt.Sprintf(
				"Failed saving configuration after getting a server: %s",
				GetErrorTraceback(saveErr),
			),
		)
	}

	return config, configType, nil
}

// SetSecureLocation sets the location for the current secure location server. countryCode is the secure location to be chosen.
// This function returns an error e.g. if the server cannot be found or the location is wrong.
func (client *Client) SetSecureLocation(countryCode string) error {
	errorMessage := "failed asking secure location"

	server, serverErr := client.Discovery.GetServerByCountryCode(countryCode, "secure_internet")
	if serverErr != nil {
		client.Logger.Error(
			fmt.Sprintf(
				"Failed getting secure internet server by country code: %s with error: %s",
				countryCode,
				GetErrorTraceback(serverErr),
			),
		)
		client.goBackInternal()
		return &types.WrappedErrorMessage{Message: errorMessage, Err: serverErr}
	}

	setLocationErr := client.Servers.SetSecureLocation(server)
	if setLocationErr != nil {
		client.Logger.Error(
			fmt.Sprintf(
				"Failed setting secure internet server with error: %s",
				GetErrorTraceback(setLocationErr),
			),
		)
		client.goBackInternal()
		return &types.WrappedErrorMessage{Message: errorMessage, Err: setLocationErr}
	}
	return nil
}

// askProfile asks the user for a profile by moving the FSM to the ASK_PROFILE state.
func (client *Client) askProfile(chosenServer server.Server) error {
	base, baseErr := chosenServer.GetBase()
	if baseErr != nil {
		return &types.WrappedErrorMessage{Message: "failed asking for profiles", Err: baseErr}
	}
	client.FSM.GoTransitionWithData(STATE_ASK_PROFILE, &base.Profiles, false)
	return nil
}

// askSecureLocation asks the user to choose a Secure Internet location by moving the FSM to the STATE_ASK_LOCATION state.
func (client *Client) askSecureLocation() error {
	locations := client.Discovery.GetSecureLocationList()

	// Ask for the location in the callback
	client.FSM.GoTransitionWithData(STATE_ASK_LOCATION, locations, false)

	// The state has changed, meaning setting the secure location was not successful
	if client.FSM.Current != STATE_ASK_LOCATION {
		// TODO: maybe a custom type for this errors.new?
		return &types.WrappedErrorMessage{
			Message: "failed setting secure location",
			Err:     errors.New("failed loading secure location"),
		}
	}
	return nil
}

// addSecureInternetHomeServer adds a Secure Internet Home Server with `orgID` that was obtained from the Discovery file.
// Because there is only one Secure Internet Home Server, it replaces the existing one.
func (client *Client) addSecureInternetHomeServer(orgID string) (server.Server, error) {
	errorMessage := fmt.Sprintf(
		"failed adding Secure Internet home server with organization ID %s",
		orgID,
	)
	// Get the secure internet URL from discovery
	secureOrg, secureServer, discoErr := client.Discovery.GetSecureHomeArgs(orgID)
	if discoErr != nil {
		return nil, &types.WrappedErrorMessage{Message: errorMessage, Err: discoErr}
	}

	// Add the secure internet server
	server, serverErr := client.Servers.AddSecureInternet(secureOrg, secureServer)

	if serverErr != nil {
		return nil, &types.WrappedErrorMessage{Message: errorMessage, Err: serverErr}
	}

	var locationErr error

	if !client.Servers.HasSecureLocation() {
		locationErr = client.askSecureLocation()
	} else {
		// reinitialize
		locationErr = client.SetSecureLocation(client.Servers.GetSecureLocation())
	}

	if locationErr != nil {
		return nil, &types.WrappedErrorMessage{Message: errorMessage, Err: locationErr}
	}

	return server, nil
}

// RemoveSecureInternet removes the current secure internet server.
// It returns an error if the server cannot be removed due to the state being DEREGISTERED.
// Note that if the server does not exist, it returns nil as an error.
func (client *Client) RemoveSecureInternet() error {
	if client.InFSMState(STATE_DEREGISTERED) {
		client.Logger.Error("Failed removing secure internet server due to deregistered")
		return &types.WrappedErrorMessage{
			Message: "failed to remove Secure Internet",
			Err:     FSMDeregisteredError{}.CustomError(),
		}
	}
	// No error because we can only have one secure internet server and if there are no secure internet servers, this is a NO-OP
	client.Servers.RemoveSecureInternet()
	client.FSM.GoTransitionWithData(STATE_NO_SERVER, client.Servers, false)
	// Save the config
	saveErr := client.Config.Save(&client)
	if saveErr != nil {
		client.Logger.Info(
			fmt.Sprintf(
				"Failed saving configuration after removing a secure internet server: %s",
				GetErrorTraceback(saveErr),
			),
		)
	}
	return nil
}

// RemoveInstituteAccess removes the institute access server with `url`.
// It returns an error if the server cannot be removed due to the state being DEREGISTERED.
// Note that if the server does not exist, it returns nil as an error.
func (client *Client) RemoveInstituteAccess(url string) error {
	if client.InFSMState(STATE_DEREGISTERED) {
		return &types.WrappedErrorMessage{
			Message: "failed to remove Institute Access",
			Err:     FSMDeregisteredError{}.CustomError(),
		}
	}
	// No error because this is a NO-OP if the server doesn't exist
	client.Servers.RemoveInstituteAccess(url)
	client.FSM.GoTransitionWithData(STATE_NO_SERVER, client.Servers, false)
	// Save the config
	saveErr := client.Config.Save(&client)
	if saveErr != nil {
		client.Logger.Info(
			fmt.Sprintf(
				"Failed saving configuration after removing an institute access server: %s",
				GetErrorTraceback(saveErr),
			),
		)
	}
	return nil
}

// RemoveCustomServer removes the custom server with `url`.
// It returns an error if the server cannot be removed due to the state being DEREGISTERED.
// Note that if the server does not exist, it returns nil as an error.
func (client *Client) RemoveCustomServer(url string) error {
	if client.InFSMState(STATE_DEREGISTERED) {
		return &types.WrappedErrorMessage{
			Message: "failed to remove Custom Server",
			Err:     FSMDeregisteredError{}.CustomError(),
		}
	}
	// No error because this is a NO-OP if the server doesn't exist
	client.Servers.RemoveCustomServer(url)
	client.FSM.GoTransitionWithData(STATE_NO_SERVER, client.Servers, false)
	// Save the config
	saveErr := client.Config.Save(&client)
	if saveErr != nil {
		client.Logger.Info(
			fmt.Sprintf(
				"Failed saving configuration after removing a custom server: %s",
				GetErrorTraceback(saveErr),
			),
		)
	}
	return nil
}

// GetConfigSecureInternet gets a configuration for a Secure Internet Server.
// It ensures that the Secure Internet Server exists by creating or using an existing one with the orgID.
// `preferTCP` indicates that the client wants to use TCP (through OpenVPN) to establish the VPN tunnel.
func (client *Client) GetConfigSecureInternet(
	orgID string,
	preferTCP bool,
) (string, string, error) {
	errorMessage := fmt.Sprintf(
		"failed getting a configuration for Secure Internet organization %s",
		orgID,
	)
	client.FSM.GoTransition(STATE_LOADING_SERVER)
	server, serverErr := client.addSecureInternetHomeServer(orgID)
	if serverErr != nil {
		client.Logger.Error(
			fmt.Sprintf(
				"Failed adding a secure internet server with error: %s",
				GetErrorTraceback(serverErr),
			),
		)
		client.goBackInternal()
		return "", "", &types.WrappedErrorMessage{Message: errorMessage, Err: serverErr}
	}

	client.FSM.GoTransition(STATE_CHOSEN_SERVER)

	config, configType, configErr := client.getConfig(server, preferTCP)
	if configErr != nil {
		client.Logger.Inherit(
			configErr,
			fmt.Sprintf(
				"Failed getting a secure internet configuration with error: %s",
				GetErrorTraceback(configErr),
			),
		)
		return "", "", &types.WrappedErrorMessage{Level: GetErrorLevel(configErr), Message: errorMessage, Err: configErr}
	}
	return config, configType, nil
}

// addInstituteServer adds an Institute Access server by `url`.
func (client *Client) addInstituteServer(url string) (server.Server, error) {
	errorMessage := fmt.Sprintf("failed adding Institute Access server with url %s", url)
	instituteServer, discoErr := client.Discovery.GetServerByURL(url, "institute_access")
	if discoErr != nil {
		return nil, &types.WrappedErrorMessage{Message: errorMessage, Err: discoErr}
	}
	// Add the secure internet server
	server, serverErr := client.Servers.AddInstituteAccessServer(instituteServer)
	if serverErr != nil {
		return nil, &types.WrappedErrorMessage{Message: errorMessage, Err: serverErr}
	}

	client.FSM.GoTransition(STATE_CHOSEN_SERVER)

	return server, nil
}

// addCustomServer adds a Custom Server by `url`
func (client *Client) addCustomServer(url string) (server.Server, error) {
	errorMessage := fmt.Sprintf("failed adding Custom server with url %s", url)

	url, urlErr := util.EnsureValidURL(url)
	if urlErr != nil {
		return nil, &types.WrappedErrorMessage{Message: errorMessage, Err: urlErr}
	}

	customServer := &types.DiscoveryServer{
		BaseURL:     url,
		DisplayName: map[string]string{"en": url},
		Type:        "custom_server",
	}

	// A custom server is just an institute access server under the hood
	server, serverErr := client.Servers.AddCustomServer(customServer)
	if serverErr != nil {
		return nil, &types.WrappedErrorMessage{Message: errorMessage, Err: serverErr}
	}

	client.FSM.GoTransition(STATE_CHOSEN_SERVER)

	return server, nil
}

// GetConfigInstituteAccess gets a configuration for an Institute Access Server.
// It ensures that the Institute Access Server exists by creating or using an existing one with the url.
// `preferTCP` indicates that the client wants to use TCP (through OpenVPN) to establish the VPN tunnel.
func (client *Client) GetConfigInstituteAccess(url string, preferTCP bool) (string, string, error) {
	errorMessage := fmt.Sprintf("failed getting a configuration for Institute Access %s", url)
	client.FSM.GoTransition(STATE_LOADING_SERVER)
	server, serverErr := client.addInstituteServer(url)
	if serverErr != nil {
		client.Logger.Error(
			fmt.Sprintf(
				"Failed adding an institute access server with error: %s",
				GetErrorTraceback(serverErr),
			),
		)
		client.goBackInternal()
		return "", "", &types.WrappedErrorMessage{Message: errorMessage, Err: serverErr}
	}

	config, configType, configErr := client.getConfig(server, preferTCP)
	if configErr != nil {
		client.Logger.Inherit(configErr,
			fmt.Sprintf(
				"Failed getting an institute access server configuration with error: %s",
				GetErrorTraceback(configErr),
			),
		)
		return "", "", &types.WrappedErrorMessage{Level: GetErrorLevel(configErr), Message: errorMessage, Err: configErr}
	}
	return config, configType, nil
}

// GetConfigCustomServer gets a configuration for a Custom Server.
// It ensures that the Custom Server exists by creating or using an existing one with the url.
// `preferTCP` indicates that the client wants to use TCP (through OpenVPN) to establish the VPN tunnel.
func (client *Client) GetConfigCustomServer(url string, preferTCP bool) (string, string, error) {
	errorMessage := fmt.Sprintf("failed getting a configuration for custom server %s", url)
	client.FSM.GoTransition(STATE_LOADING_SERVER)
	server, serverErr := client.addCustomServer(url)

	if serverErr != nil {
		client.Logger.Error(
			fmt.Sprintf(
				"Failed adding a custom server with error: %s",
				GetErrorTraceback(serverErr),
			),
		)
		client.goBackInternal()
		return "", "", &types.WrappedErrorMessage{Message: errorMessage, Err: serverErr}
	}

	config, configType, configErr := client.getConfig(server, preferTCP)
	if configErr != nil {
		client.Logger.Inherit(
			configErr,
			fmt.Sprintf(
				"Failed getting a custom server with error: %s",
				GetErrorTraceback(configErr),
			),
		)
		return "", "", &types.WrappedErrorMessage{Level: GetErrorLevel(configErr), Message: errorMessage, Err: configErr}
	}
	return config, configType, nil
}

// CancelOAuth cancels OAuth if one is in progress.
// If OAuth is not in progress, it returns an error.
// An error is also returned if OAuth is in progress but it fails to cancel it.
func (client *Client) CancelOAuth() error {
	errorMessage := "failed to cancel OAuth"
	if !client.InFSMState(STATE_OAUTH_STARTED) {
		client.Logger.Error("Failed cancelling OAuth, not in the right state")
		return &types.WrappedErrorMessage{
			Message: errorMessage,
			Err: FSMWrongStateError{
				Got:  client.FSM.Current,
				Want: STATE_OAUTH_STARTED,
			}.CustomError(),
		}
	}

	currentServer, serverErr := client.Servers.GetCurrentServer()
	if serverErr != nil {
		client.Logger.Warning(
			fmt.Sprintf(
				"Failed cancelling OAuth, no server configured to cancel OAuth for (err: %v)",
				serverErr,
			),
		)
		return &types.WrappedErrorMessage{Message: errorMessage, Err: serverErr}
	}
	server.CancelOAuth(currentServer)
	return nil
}

// ChangeSecureLocation changes the location for an existing Secure Internet Server.
// Changing a secure internet location is only possible when the user is in the main screen (STATE_NO_SERVER), otherwise it returns an error.
// It also returns an error if something has gone wrong when selecting the new location
func (client *Client) ChangeSecureLocation() error {
	errorMessage := "failed to change location from the main screen"

	if !client.InFSMState(STATE_NO_SERVER) {
		client.Logger.Error("Failed changing secure internet location, not in the right state")
		return &types.WrappedErrorMessage{
			Message: errorMessage,
			Err: FSMWrongStateError{
				Got:  client.FSM.Current,
				Want: STATE_NO_SERVER,
			}.CustomError(),
		}
	}

	askLocationErr := client.askSecureLocation()
	if askLocationErr != nil {
		client.Logger.Error(
			fmt.Sprintf(
				"Failed changing secure internet location, err: %s",
				GetErrorTraceback(askLocationErr),
			),
		)
		return &types.WrappedErrorMessage{Message: errorMessage, Err: askLocationErr}
	}

	// Go back to the main screen
	client.FSM.GoTransitionWithData(STATE_NO_SERVER, client.Servers, false)

	return nil
}

// GetDiscoOrganizations gets the organizations list from the discovery server
// If the list cannot be retrieved an error is returned.
// If this is the case then a previous version of the list is returned if there is any.
// This takes into account the frequency of updates, see: https://github.com/eduvpn/documentation/blob/v3/SERVER_DISCOVERY.md#organization-list.
func (client *Client) GetDiscoOrganizations() (*types.DiscoveryOrganizations, error) {
	orgs, orgsErr := client.Discovery.GetOrganizationsList()
	if orgsErr != nil {
		client.Logger.Warning(
			fmt.Sprintf(
				"Failed getting discovery organizations, Err: %s",
				GetErrorTraceback(orgsErr),
			),
		)
		return nil, &types.WrappedErrorMessage{
			Message: "failed getting discovery organizations list",
			Err:     orgsErr,
		}
	}
	return orgs, nil
}

// GetDiscoDiscovers gets the servers list from the discovery server
// If the list cannot be retrieved an error is returned.
// If this is the case then a previous version of the list is returned if there is any.
// This takes into account the frequency of updates, see: https://github.com/eduvpn/documentation/blob/v3/SERVER_DISCOVERY.md#server-list.
func (client *Client) GetDiscoServers() (*types.DiscoveryServers, error) {
	servers, serversErr := client.Discovery.GetServersList()
	if serversErr != nil {
		client.Logger.Warning(
			fmt.Sprintf("Failed getting discovery servers, Err: %s", GetErrorTraceback(serversErr)),
		)
		return nil, &types.WrappedErrorMessage{
			Message: "failed getting discovery servers list",
			Err:     serversErr,
		}
	}
	return servers, nil
}

// SetProfileID sets a `profileID` for the current server.
// An error is returned if this is not possible, for example when no server is configured.
func (client *Client) SetProfileID(profileID string) error {
	errorMessage := "failed to set the profile ID for the current server"
	server, serverErr := client.Servers.GetCurrentServer()
	if serverErr != nil {
		client.Logger.Warning(
			fmt.Sprintf(
				"Failed setting a profile ID because no server configured, Err: %s",
				GetErrorTraceback(serverErr),
			),
		)
		client.goBackInternal()
		return &types.WrappedErrorMessage{Message: errorMessage, Err: serverErr}
	}

	base, baseErr := server.GetBase()
	if baseErr != nil {
		client.Logger.Error(
			fmt.Sprintf("Failed setting a profile ID, Err: %s", GetErrorTraceback(serverErr)),
		)
		client.goBackInternal()
		return &types.WrappedErrorMessage{Message: errorMessage, Err: baseErr}
	}
	base.Profiles.Current = profileID
	return nil
}

// SetSearchServer sets the FSM to the SEARCH_SERVER state.
// This indicates that the user wants to search for a new server.
// Returns an error if this state transition is not possible.
func (client *Client) SetSearchServer() error {
	if !client.FSM.HasTransition(STATE_SEARCH_SERVER) {
		client.Logger.Warning(
			fmt.Sprintf(
				"Failed setting search server, wrong state %s",
				GetStateName(client.FSM.Current),
			),
		)
		return &types.WrappedErrorMessage{
			Message: "failed to set search server",
			Err: FSMWrongStateTransitionError{
				Got:  client.FSM.Current,
				Want: STATE_CONNECTED,
			}.CustomError(),
		}
	}

	client.FSM.GoTransition(STATE_SEARCH_SERVER)
	return nil
}

// SetConnected sets the FSM to the CONNECTED state.
// This indicates that the VPN is connected to the server.
// Returns an error if this state transition is not possible.
func (client *Client) SetConnected() error {
	errorMessage := "failed to set connected"
	if client.InFSMState(STATE_CONNECTED) {
		// already connected, show no error
		client.Logger.Warning("Already connected")
		return nil
	}
	if !client.FSM.HasTransition(STATE_CONNECTED) {
		client.Logger.Warning(
			fmt.Sprintf(
				"Failed setting connected, wrong state: %s",
				GetStateName(client.FSM.Current),
			),
		)
		return &types.WrappedErrorMessage{
			Message: errorMessage,
			Err: FSMWrongStateTransitionError{
				Got:  client.FSM.Current,
				Want: STATE_CONNECTED,
			}.CustomError(),
		}
	}

	currentServer, currentServerErr := client.Servers.GetCurrentServer()
	if currentServerErr != nil {
		client.Logger.Warning(
			fmt.Sprintf(
				"Failed setting connected, cannot get current server with error: %s",
				GetErrorTraceback(currentServerErr),
			),
		)
		return &types.WrappedErrorMessage{Message: errorMessage, Err: currentServerErr}
	}

	client.FSM.GoTransitionWithData(STATE_CONNECTED, currentServer, false)
	return nil
}

// SetConnecting sets the FSM to the CONNECTING state.
// This indicates that the VPN is currently connecting to the server.
// Returns an error if this state transition is not possible.
func (client *Client) SetConnecting() error {
	errorMessage := "failed to set connecting"
	if client.InFSMState(STATE_CONNECTING) {
		// already loading connection, show no error
		client.Logger.Warning("Already connecting")
		return nil
	}
	if !client.FSM.HasTransition(STATE_CONNECTING) {
		client.Logger.Warning(
			fmt.Sprintf(
				"Failed setting connecting, wrong state: %s",
				GetStateName(client.FSM.Current),
			),
		)
		return &types.WrappedErrorMessage{
			Message: errorMessage,
			Err: FSMWrongStateTransitionError{
				Got:  client.FSM.Current,
				Want: STATE_CONNECTING,
			}.CustomError(),
		}
	}

	currentServer, currentServerErr := client.Servers.GetCurrentServer()
	if currentServerErr != nil {
		client.Logger.Warning(
			fmt.Sprintf(
				"Failed setting connecting, cannot get current server with error: %s",
				GetErrorTraceback(currentServerErr),
			),
		)
		return &types.WrappedErrorMessage{Message: errorMessage, Err: currentServerErr}
	}

	client.FSM.GoTransitionWithData(STATE_CONNECTING, currentServer, false)
	return nil
}

// SetDisconnecting sets the FSM to the DISCONNECTING state.
// This indicates that the VPN is currently disconnecting from the server.
// Returns an error if this state transition is not possible.
func (client *Client) SetDisconnecting() error {
	errorMessage := "failed to set disconnecting"
	if client.InFSMState(STATE_DISCONNECTING) {
		// already disconnecting, show no error
		client.Logger.Warning("Already disconnecting")
		return nil
	}
	if !client.FSM.HasTransition(STATE_DISCONNECTING) {
		client.Logger.Warning(
			fmt.Sprintf(
				"Failed setting disconnecting, wrong state: %s",
				GetStateName(client.FSM.Current),
			),
		)
		return &types.WrappedErrorMessage{
			Message: errorMessage,
			Err: FSMWrongStateTransitionError{
				Got:  client.FSM.Current,
				Want: STATE_DISCONNECTING,
			}.CustomError(),
		}
	}

	currentServer, currentServerErr := client.Servers.GetCurrentServer()
	if currentServerErr != nil {
		client.Logger.Warning(
			fmt.Sprintf(
				"Failed setting disconnected, cannot get current server with error: %s",
				GetErrorTraceback(currentServerErr),
			),
		)
		return &types.WrappedErrorMessage{Message: errorMessage, Err: currentServerErr}
	}

	client.FSM.GoTransitionWithData(STATE_DISCONNECTING, currentServer, false)
	return nil
}

// SetDisconnected sets the FSM to the DISCONNECTED state.
// This indicates that the VPN is currently disconnected from the server.
// This also sends the /disconnect API call to the server.
// Returns an error if this state transition is not possible.
func (client *Client) SetDisconnected(cleanup bool) error {
	errorMessage := "failed to set disconnected"
	if client.InFSMState(STATE_DISCONNECTED) {
		// already disconnected, show no error
		client.Logger.Warning("Already disconnected")
		return nil
	}
	if !client.FSM.HasTransition(STATE_DISCONNECTED) {
		client.Logger.Warning(
			fmt.Sprintf(
				"Failed setting disconnected, wrong state: %s",
				GetStateName(client.FSM.Current),
			),
		)
		return &types.WrappedErrorMessage{
			Message: errorMessage,
			Err: FSMWrongStateTransitionError{
				Got:  client.FSM.Current,
				Want: STATE_DISCONNECTED,
			}.CustomError(),
		}
	}

	currentServer, currentServerErr := client.Servers.GetCurrentServer()
	if currentServerErr != nil {
		client.Logger.Warning(
			fmt.Sprintf(
				"Failed setting disconnect, failed getting current server with error: %s",
				GetErrorTraceback(currentServerErr),
			),
		)
		return &types.WrappedErrorMessage{Message: errorMessage, Err: currentServerErr}
	}

	if cleanup {
		// Do the /disconnect API call and go to disconnected after...
		server.Disconnect(currentServer)
	}

	client.FSM.GoTransitionWithData(STATE_DISCONNECTED, currentServer, false)

	return nil
}

// RenewSession renews the session for the current VPN server.
// This logs the user back in.
func (client *Client) RenewSession() error {
	errorMessage := "failed to renew session"

	currentServer, currentServerErr := client.Servers.GetCurrentServer()
	if currentServerErr != nil {
		client.Logger.Warning(
			fmt.Sprintf(
				"Failed getting current server to renew, error: %s",
				GetErrorTraceback(currentServerErr),
			),
		)
		return &types.WrappedErrorMessage{Message: errorMessage, Err: currentServerErr}
	}

	loginErr := client.ensureLogin(currentServer)
	if loginErr != nil {
		client.Logger.Warning(
			fmt.Sprintf(
				"Failed logging in server for renew, error: %s",
				GetErrorTraceback(loginErr),
			),
		)
		return &types.WrappedErrorMessage{Message: errorMessage, Err: loginErr}
	}

	return nil
}

// ShouldRenewButton returns true if the renew button should be shown
// If there is no server then this returns false and logs with INFO if so
// In other cases it simply checks the expiry time and calculates according to: https://github.com/eduvpn/documentation/blob/b93854dcdd22050d5f23e401619e0165cb8bc591/API.md#session-expiry.
func (client *Client) ShouldRenewButton() bool {
	if !client.InFSMState(STATE_CONNECTED) && !client.InFSMState(STATE_CONNECTING) &&
		!client.InFSMState(STATE_DISCONNECTED) &&
		!client.InFSMState(STATE_DISCONNECTING) {
		return false
	}

	currentServer, currentServerErr := client.Servers.GetCurrentServer()

	if currentServerErr != nil {
		client.Logger.Info(
			fmt.Sprintf(
				"No server found to renew with err: %s",
				GetErrorTraceback(currentServerErr),
			),
		)
		return false
	}

	return server.ShouldRenewButton(currentServer)
}

// InFSMState is a helper to check if the FSM is in state `checkState`.
func (client *Client) InFSMState(checkState FSMStateID) bool {
	return client.FSM.InState(checkState)
}

// GetErrorCause gets the cause for error `err`.
func GetErrorCause(err error) error {
	return types.GetErrorCause(err)
}

// GetErrorCause gets the level for error `err`.
func GetErrorLevel(err error) types.ErrorLevel {
	return types.GetErrorLevel(err)
}

// GetErrorCause gets the traceback for error `err`.
func GetErrorTraceback(err error) string {
	return types.GetErrorTraceback(err)
}

// GetTranslated gets the translation for `languages` using the current state language.
func (client *Client) GetTranslated(languages map[string]string) string {
	return util.GetLanguageMatched(languages, client.Language)
}