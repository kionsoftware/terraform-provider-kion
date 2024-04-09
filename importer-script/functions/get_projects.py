from .api_call import api_call
from constants import BASE_URL


def get_projects(cloud_rule=False):
    """
    Get Projects

    Params:
        cloud_rule (dict) - cloud rule object for which to return project IDs where the cloud rule is applied locally
                            if not set, it will return a list of all projects from Kion
    Return:
        success - list of project IDs or project objects (based on cloud_rule param)
        failure - False
    """
    if cloud_rule:
        ids = []

        # pull out the ID of the provided cloud rule
        c_id = cloud_rule['cloud_rule']['id']

        # for each project, call the v3/project/{id}/cloud-rule endpoint
        # to get it's locally applied cloud rules. Then check if
        # c_id is in that list of cloud rules
        for p in cloud_rule['projects']:
            p_id = p['id']
            url = "%s/v3/project/%s/cloud-rule" % (BASE_URL, p_id)
            project_rules = api_call(url)
            if project_rules:
                for rule in project_rules:
                    if rule['id'] == c_id:
                        if p_id not in ids:
                            ids.append(p_id)
        return ids
    else:
        url = '%s/v3/project' % BASE_URL
        projects = api_call(url)
        if projects:
            return projects
        else:
            print("Could not get projects from Kion.")
            return False
