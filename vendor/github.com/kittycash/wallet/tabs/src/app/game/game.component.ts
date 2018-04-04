import { Component, OnInit, Directive, ElementRef, HostListener, HostBinding, Renderer, Input } from '@angular/core';
import { ScoreboardService } from './scoreboard.service';
import { finalize } from 'rxjs/operators';
import { environment } from '../../environments/environment';

@Component({
  selector: 'game',
  templateUrl: './game.component.html',
  styleUrls: ['./game.component.scss']
})
export class GameComponent implements OnInit {
 
  gameUrl: string;
  scores: any;
  span: string;
  isLoading: boolean;

  constructor(private scoreboardService: ScoreboardService) { 
  	this.gameUrl = environment.serverUrl + "/scoreboard/game";
  }

 
  ngOnInit() {
    this.loadScores('day');
  }

  loadScores(span: string) {
  	this.span = span;
  	this.isLoading = true;
    this.scoreboardService.getScores({span: this.span})
      .pipe(finalize(() => { this.isLoading = false; }))
      .subscribe((scores: any) => { this.scores = scores; });
  }
}
