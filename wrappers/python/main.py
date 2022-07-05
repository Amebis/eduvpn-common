import eduvpncommon.main as eduvpn
import webbrowser
import json

# Asks the user for a profile index
# It loops up until a valid input is given
def ask_profile_input(total: int) -> int:
    profile_index = None

    while profile_index is None:
        try:
            profile_index = int(
                input("Please select a profile by inputting a number (e.g. 1): ")
            )
            if (profile_index > total) or (profile_index < 1):
                print("Invalid profile range")
                profile_index = None
        except ValueError:
            print("Please enter a valid input")

    # The profile is one based, move to zero based input
    return profile_index - 1


# Sets up the callbacks using the provided class
def setup_callbacks(_eduvpn: eduvpn.EduVPN) -> None:
    # The callback that starst OAuth
    # It needs to open the URL in the web browser
    @_eduvpn.event.on("OAuth_Started", eduvpn.StateType.Enter)
    def oauth_initialized(old_state: str, url: str) -> None:
        print(f"Got OAuth URL {url}, old state: {old_state}")
        webbrowser.open(url)

    # The callback which asks the user for a profile
    @_eduvpn.event.on("Ask_Profile", eduvpn.StateType.Enter)
    def ask_profile(old_state: str, profiles: str):
        print(
            "Multiple profiles found, you need to select a profile, old state: {old_state}"
        )

        # Parse the profiles as JSON
        data = json.loads(profiles)

        # Get a lits of profiles
        profile_strings = [x["profile_id"] for x in data["info"]["profile_list"]]
        total_profiles = len(profile_strings)

        # Create a list of the strings to standard output
        for idx, profile in enumerate(profile_strings):
            print(f"{idx+1}. {profile}")

        # Get the profile index from the user
        profile_index = ask_profile_input(total_profiles)

        # Set the profile with the index
        _eduvpn.set_profile(profile_strings[profile_index])


# The main entry point
if __name__ == "__main__":
    _eduvpn = eduvpn.EduVPN("org.eduvpn.app.linux", "configs")
    setup_callbacks(_eduvpn)

    # Register with the eduVPN-common library
    try:
        _eduvpn.register(debug=True)
    except Exception as e:
        print("Failed registering:", e)

    server = input(
        "Which server (Custom/Institute Access) do you want to connect to? (e.g. https://eduvpn.example.com): "
    )

    # Ensure we have a valid http prefix
    if not server.startswith("http"):
        # https by default
        server = "https://" + server

    # Get a Wireguard/OpenVPN config
    try:
        config, config_type = _eduvpn.get_config_custom_server(server)
    except Exception as e:
        print("Failed to connect:", e)
    print(f"Got a config with type: {config_type} and contents:\n{config}")

    # Set the internal FSM state to connected
    try:
        _eduvpn.set_connected()
    except Exception as e:
        print("Failed to set connected:", e)

    # Save and exit
    _eduvpn.deregister()
