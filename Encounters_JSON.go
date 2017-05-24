package Encounters_JSON

type Encounter struct{
  ResourceType string
  Id string
  Meta struct{
    VersionId string
    LastUpdated string
  }
  Status string
  Type []struct{
    Text string
  }
  Patient struct{
    Reference string
  }
  Participant []struct{
    Individual struct{
      Reference string
    }
  }
  Period struct{
    Start string
  }
  Reason []struct{
    Text string
  }
  Indication []struct{
    Reference string
  }
}
