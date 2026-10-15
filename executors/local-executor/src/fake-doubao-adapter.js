class FakeDoubaoAdapter {
  constructor(script = {}) {
    this.script = script;
    this.calls = [];
  }

  async prepare(input) {
    this.calls.push(['prepare', input.account.id]);
  }

  async submit() {
    this.calls.push(['submit']);
    return this.script.submit || { status: 'accepted', submissionId: 'fake-submission' };
  }

  async recoverAcceptance() {
    this.calls.push(['recoverAcceptance']);
    return this.script.recover || { status: 'accepted', submissionId: 'fake-submission' };
  }

  async waitForCompletion() {
    this.calls.push(['waitForCompletion']);
    return this.script.completion || { mediaId: 'fake-media' };
  }

  async fetchArtifact() {
    this.calls.push(['fetchArtifact']);
    return this.script.artifact || { artifactId: 'fake-artifact' };
  }
}

module.exports = { FakeDoubaoAdapter };
