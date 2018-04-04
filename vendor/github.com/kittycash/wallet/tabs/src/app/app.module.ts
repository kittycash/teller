import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { HttpClientModule } from '@angular/common/http';
import { MarketplaceAppModule } from 'kittycash-marketplace-lib';
import { GameComponent } from "./game/game.component";
import { ScoreboardService } from "./game/scoreboard.service";
import { AppComponent } from './app.component';
import { SafePipe } from './game/safe.pipe';

@NgModule({
  declarations: [
    AppComponent,
    GameComponent,
    SafePipe
  ],
  imports: [
  	HttpClientModule,
    BrowserModule,
    MarketplaceAppModule
  ],
  providers: [
  	ScoreboardService
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
