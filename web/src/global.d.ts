declare global {
  var Go: new () => { importObject: WebAssembly.Imports; run(instance: WebAssembly.Instance): Promise<void> };
  var mktax: { parse(input: string): string; calculate(input: string): string; version(): string } | undefined;
}

export {};
