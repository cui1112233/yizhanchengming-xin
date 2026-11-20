/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */
longlong * FUN_1800cd2d0(longlong param_1,undefined8 param_2,undefined8 *param_3)
{
  int *piVar1;
  undefined8 uVar2;
  int iVar3;
  longlong *plVar4;
  longlong *plVar5;
  longlong *plVar6;
  longlong *plVar7;
  longlong *plVar8;
  code *pcVar9;
  code *pcVar10;
  undefined8 uVar11;
  undefined8 uVar12;
  undefined8 uVar13;
  longlong *plVar14;
  longlong *plVar15;
  longlong lVar16;
  undefined8 uVar17;
  undefined1 auStack_3b8 [32];
  longlong *local_398;
  longlong *local_390;
  longlong *local_388;
  longlong *local_380;
  longlong *local_378;
  longlong *local_370;
  longlong *local_368;
  longlong *local_360;
  longlong *local_358;
  longlong *local_350;
  longlong *local_348;
  longlong *local_340;
  code *local_338;
  longlong *local_330;
  longlong *local_328;
  longlong *local_320;
  longlong *local_318;
  longlong *local_310;
  longlong *local_308;
  longlong *local_300;
  longlong *local_2f8;
  longlong *local_2f0;
  longlong *local_2e8;
  longlong *local_2e0;
  longlong local_2d8;
  longlong *local_2d0;
  longlong *local_2c8;
  longlong *local_2c0;
  longlong *local_2b8;
  longlong *local_2b0;
  longlong *local_2a8;
  longlong *local_2a0;
  longlong *local_298;
  longlong local_290;
  longlong *local_288;
  longlong *local_280;
  longlong *local_278;
  longlong *local_270;
  longlong *local_268;
  longlong *local_260;
  longlong *local_258;
  longlong *local_250;
  longlong *local_248;
  longlong *local_240;
  longlong *local_238;
  longlong *local_230;
  longlong *local_228;
  longlong *local_220;
  code *local_218;
  longlong *local_210;
  longlong *local_208;
  longlong *local_200;
  longlong *local_1f8;
  longlong *local_1f0;
  longlong *local_1e8;
  longlong *local_1e0;
  longlong *local_1d8;
  longlong *local_1d0;
  longlong *local_1c8;
  longlong *local_1c0;
  longlong *local_1b8;
  longlong *local_1b0;
  longlong *local_1a8;
  longlong *local_1a0;
  undefined8 local_198;
  longlong *local_190;
  longlong *local_188;
  undefined8 local_180;
  longlong *local_178;
  longlong *local_170;
  longlong *local_168;
  undefined8 local_160;
  longlong *local_158;
  longlong *local_150;
  longlong *local_148;
  longlong *local_140;
  longlong *local_138;
  longlong *local_130;
  longlong *local_128;
  undefined8 local_120;
  longlong *local_118;
  longlong *local_110;
  longlong *local_108;
  longlong *local_100;
  longlong *local_f8;
  longlong *local_f0;
  longlong *local_e8;
  longlong *local_e0;
  longlong *local_d8;
  longlong *local_d0;
  longlong *local_c8;
  code *local_c0;
  undefined8 local_b8;
  code *local_b0;
  longlong *local_a8;
  longlong *local_a0;
  longlong *local_98;
  longlong *local_90;
  longlong *local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  longlong *local_48;
  ulonglong local_40;
  plVar15 = DAT_1803b29f0;
  local_40 = DAT_1803a6040 ^ (ulonglong)auStack_3b8;
  local_230 = (longlong *)param_3[1];
  plVar8 = (longlong *)*param_3;
  plVar6 = (longlong *)param_3[6];
  plVar5 = (longlong *)param_3[8];
  local_268 = (longlong *)param_3[2];
  local_228 = (longlong *)param_3[3];
  local_210 = (longlong *)param_3[4];
  local_1e8 = (longlong *)param_3[5];
  local_278 = (longlong *)param_3[7];
  local_260 = (longlong *)param_3[9];
  local_1e0 = (longlong *)0x0;
  local_1d8 = (longlong *)0x0;
  local_2d0 = (longlong *)0x0;
  local_1c8 = (longlong *)0x0;
  local_218 = (code *)0x0;
  local_258 = (longlong *)0x0;
  local_250 = (longlong *)0x0;
  local_248 = (longlong *)0x0;
  local_1c0 = (longlong *)0x0;
  local_1b8 = (longlong *)0x0;
  local_1b0 = (longlong *)0x0;
  local_1a8 = (longlong *)0x0;
  local_1f8 = (longlong *)0x0;
  local_1f0 = (longlong *)0x0;
  local_240 = (longlong *)0x0;
  local_238 = (longlong *)0x0;
  local_2d8 = 0;
  local_2b0 = DAT_1803b29f0;
  local_2c8 = plVar5;
  local_2b8 = plVar8;
  local_220 = plVar6;
  if (DAT_1803b29f0 == (longlong *)0x0) {
LAB_1800cd3f2:
    DAT_1803b29f0 = (longlong *)FUN_18036cd10(param_1,DAT_1803b2918,DAT_1803ca7a0,0xd0);
  }
  else if ((1 < *DAT_1803b29f0) || (DAT_1803b29f0[2] != 0)) {
    if ((DAT_1803b29f0 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b29f0 && (*DAT_1803b29f0 = *DAT_1803b29f0 + -1, *plVar15 == 0)))) {
      (**(code **)(plVar15[1] + 0x30))(plVar15);
    }
    goto LAB_1800cd3f2;
  }
  plVar15 = DAT_1803b29f0;
  local_2b0 = DAT_1803b29f0;
  FUN_180001d00(param_1,DAT_1803b29f0);
  uVar17 = 0xb4;
  *(undefined4 *)(plVar15 + 5) = 0xb4;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#81="_check_cancel" bytes=[3953713, 3953728] */
  plVar4 = (longlong *)FUN_18037e560(param_1,plVar8,DAT_1803b2d88);
  if (plVar4 == (longlong *)0x0) {
    local_2d8 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar6 = (longlong *)0x0;
    goto LAB_1800cfeec;
  }
  if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
    (**(code **)(plVar4[1] + 0x30))();
  }
  uVar17 = 0xb9;
  local_170 = local_230;
  *(undefined4 *)(plVar15 + 5) = 0xb9;
  local_168 = plVar5;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#82="_resume_or_run_deconstruction" bytes=[3953728, 3953759] */
  plVar4 = (longlong *)FUN_18037e820(param_1,plVar8,DAT_1803b2d90,&local_170);
  local_1a0 = plVar4;
  if (plVar4 == (longlong *)0x0) {
LAB_1800cd4b5:
    local_2d8 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar5 = local_2c8;
    plVar6 = (longlong *)0x0;
  }
  else {
    uVar17 = 0xbd;
    *(undefined4 *)(plVar15 + 5) = 0xbd;
    local_1e0 = plVar4;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#81="_check_cancel" bytes=[3953713, 3953728] */
    plVar5 = (longlong *)FUN_18037e560(param_1,plVar8,DAT_1803b2d88);
    if (plVar5 == (longlong *)0x0) goto LAB_1800cd4b5;
    if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
    local_90 = local_228;
    uVar17 = 0xc2;
    local_88 = local_210;
    *(undefined4 *)(plVar15 + 5) = 0xc2;
    local_78 = local_2c8;
    local_98 = plVar4;
    local_80 = plVar6;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#83="_resume_or_run_directions" bytes=[3953759, 3953786] */
    local_188 = (longlong *)FUN_18037ec40(param_1,plVar8,DAT_1803b2d98,&local_98);
    if (local_188 == (longlong *)0x0) goto LAB_1800cd4b5;
    uVar17 = 0xc6;
    *(undefined4 *)(plVar15 + 5) = 0xc6;
    local_1d8 = local_188;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#81="_check_cancel" bytes=[3953713, 3953728] */
    plVar6 = (longlong *)FUN_18037e560(param_1,plVar8,DAT_1803b2d88);
    if (plVar6 == (longlong *)0x0) goto LAB_1800cd4b5;
    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))(plVar6);
    }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
    plVar6 = (longlong *)FUN_18037f0c0(param_1,local_260,DAT_1803b2b00);
    uVar17 = 0xcb;
    if (plVar6 == (longlong *)0x0) {
      local_2d8 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      plVar5 = local_2c8;
      plVar6 = local_2d0;
    }
    else {
      *(undefined4 *)(plVar15 + 5) = 0xcb;
      plVar7 = (longlong *)FUN_1803770d0(param_1,plVar6,DAT_1803b2da0);
      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
        (**(code **)(plVar6[1] + 0x30))(plVar6);
      }
      if (plVar7 == (longlong *)0x0) {
        local_2d8 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar5 = local_2c8;
        plVar6 = local_2d0;
      }
      else {
        iVar3 = FUN_180001510(plVar7);
        if (iVar3 == -1) {
          local_2d8 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          plVar5 = local_2c8;
          plVar6 = local_2d0;
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
            plVar5 = local_2c8;
            plVar6 = local_2d0;
          }
        }
        else {
          if (iVar3 != 1) {
            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
              (**(code **)(plVar7[1] + 0x30))(plVar7);
            }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
            plVar7 = DAT_1803b2b10;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
            iVar3 = (int)*DAT_1803b2b10 + 1;
            if (iVar3 != 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
              *(int *)DAT_1803b2b10 = iVar3;
            }
          }
          plVar6 = (longlong *)FUN_1803749a0(plVar7);
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
          if (plVar6 == (longlong *)0x0) {
            local_2d8 = *(longlong *)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar5 = local_2c8;
            plVar6 = local_2d0;
          }
          else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#3="strip" bytes=[3952481, 3952488] */
            plVar8 = (longlong *)FUN_18037f0c0(param_1,plVar6,DAT_1803b2b18);
            if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
              (**(code **)(plVar6[1] + 0x30))(plVar6);
            }
            if (plVar8 == (longlong *)0x0) {
              local_2d8 = *(longlong *)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              plVar5 = local_2c8;
              plVar6 = local_2d0;
              plVar8 = local_2b8;
            }
            else {
              *(undefined4 *)(plVar15 + 5) = 0xcb;
              plVar6 = (longlong *)FUN_180376210(param_1,plVar8);
              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                (**(code **)(plVar8[1] + 0x30))(plVar8);
              }
              if (plVar6 == (longlong *)0x0) {
                local_2d8 = *(longlong *)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar5 = local_2c8;
                plVar6 = local_2d0;
                plVar8 = local_2b8;
              }
              else {
                local_2d0 = plVar6;
                iVar3 = FUN_180001510(plVar6);
                plVar8 = local_2b8;
                if (iVar3 == -1) {
                  local_2d8 = *(longlong *)(param_1 + 0x60);
                  uVar17 = 0xcc;
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  plVar5 = local_2c8;
                }
                else {
                  plVar7 = plVar6;
                  if (iVar3 == 0) {
                    uVar17 = 0xcd;
                    *(undefined4 *)(plVar15 + 5) = 0xcd;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#85="_extract_first_direction" bytes=[3953807, 3953833] */
                    plVar7 = (longlong *)FUN_18037e6c0(param_1,local_2b8,DAT_1803b2da8,local_188);
                    if (plVar7 == (longlong *)0x0) {
                      local_2d8 = *(longlong *)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      plVar5 = local_2c8;
                      goto LAB_1800cfeec;
                    }
                    local_2d0 = plVar7;
                    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                      (**(code **)(plVar6[1] + 0x30))(plVar6);
                    }
                  }
                  plVar5 = local_2c8;
                  local_60 = local_268;
                  uVar17 = 0xcf;
                  local_58 = local_228;
                  *(undefined4 *)(plVar15 + 5) = 0xcf;
                  local_50 = local_220;
                  local_48 = local_2c8;
                  local_70 = plVar4;
                  local_68 = plVar7;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#86="_resume_or_run_blueprint" bytes=[3953833, 3953859] */
                  local_1d0 = (longlong *)FUN_18037eda0(param_1,plVar8,DAT_1803b2db0,&local_70);
                  plVar6 = plVar7;
                  if (local_1d0 != (longlong *)0x0) {
                    uVar17 = 0xd4;
                    *(undefined4 *)(plVar15 + 5) = 0xd4;
                    local_1c8 = local_1d0;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#81="_check_cancel" bytes=[3953713, 3953728] */
                    plVar4 = (longlong *)FUN_18037e560(param_1,plVar8,DAT_1803b2d88);
                    if (plVar4 != (longlong *)0x0) {
                      if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                        (**(code **)(plVar4[1] + 0x30))(plVar4);
                      }
                      lVar16 = FUN_1800c7020(param_1);
                      if (lVar16 == 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#4="FileRepository" bytes=[3952488, 3952504] */
                        FUN_180373340(param_1,&local_2d8,DAT_1803b2b20);
                        uVar17 = 0xd9;
                      }
                      else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#87="write_text" bytes=[3953859, 3953871] */
                        plVar4 = (longlong *)FUN_18037f0c0(param_1,lVar16,DAT_1803b2db8);
                        if (plVar4 == (longlong *)0x0) {
                          local_2d8 = *(longlong *)(param_1 + 0x60);
                          uVar17 = 0xd9;
                          *(undefined8 *)(param_1 + 0x60) = 0;
                        }
                        else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#50="imitation" bytes=[3953271, 3953282] */
                          plVar7 = (longlong *)FUN_180382440(plVar5,DAT_1803b2c90);
                          if (plVar7 == (longlong *)0x0) {
                            local_2d8 = *(longlong *)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                              (**(code **)(plVar4[1] + 0x30))(plVar4);
                            }
                            uVar17 = 0xd9;
                          }
                          else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#88="reference.md" bytes=[3953871, 3953885] */
                            plVar6 = (longlong *)FUN_180382440(plVar7,DAT_1803b2dc0);
                            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                            }
                            if (plVar6 == (longlong *)0x0) {
                              local_2d8 = *(longlong *)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                                (**(code **)(plVar4[1] + 0x30))(plVar4);
                              }
                              uVar17 = 0xd9;
                              plVar6 = local_2d0;
                            }
                            else {
                              uVar17 = 0xd9;
                              local_280 = local_230;
                              *(undefined4 *)(plVar15 + 5) = 0xd9;
                              local_288 = plVar6;
                              plVar8 = (longlong *)FUN_180377750(param_1,plVar4,&local_288);
                              if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                                (**(code **)(plVar4[1] + 0x30))(plVar4);
                              }
                              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                (**(code **)(plVar6[1] + 0x30))(plVar6);
                              }
                              if (plVar8 == (longlong *)0x0) {
                                local_2d8 = *(longlong *)(param_1 + 0x60);
                                *(undefined8 *)(param_1 + 0x60) = 0;
                                plVar6 = local_2d0;
                                plVar8 = local_2b8;
                              }
                              else {
                                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                                }
                                plVar8 = local_228;
                                iVar3 = FUN_180001510(local_228);
                                if (iVar3 == -1) {
                                  local_2d8 = *(longlong *)(param_1 + 0x60);
                                  uVar17 = 0xda;
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                  plVar6 = local_2d0;
                                  plVar8 = local_2b8;
                                }
                                else if (iVar3 == 0) {
LAB_1800cdbd3:
                                  plVar8 = local_210;
                                  iVar3 = FUN_180001510(local_210);
                                  if (iVar3 == -1) {
                                    FUN_18000ce30(param_1,&local_2d8);
                                    uVar17 = 0xdc;
                                    plVar6 = local_2d0;
                                    plVar8 = local_2b8;
                                  }
                                  else if (iVar3 == 0) {
LAB_1800cdd65:
                                    plVar8 = local_220;
                                    iVar3 = FUN_180001510(local_220);
                                    if (iVar3 == -1) {
                                      FUN_18000ce30(param_1,&local_2d8);
                                      uVar17 = 0xde;
                                      plVar6 = local_2d0;
                                      plVar8 = local_2b8;
                                    }
                                    else if (iVar3 == 0) {
LAB_1800cdef7:
                                      plVar8 = local_260;
                                      plVar6 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
                                               FUN_18037f0c0(param_1,local_260,DAT_1803b2b00);
                                      if (plVar6 == (longlong *)0x0) {
                                        FUN_18000ce30(param_1,&local_2d8);
                                        uVar17 = 0xe1;
                                        plVar6 = local_2d0;
                                        plVar8 = local_2b8;
                                      }
                                      else {
                                        uVar17 = 0xe1;
                                        *(undefined4 *)(plVar15 + 5) = 0xe1;
                                        pcVar9 = (code *)FUN_1803770d0(param_1,plVar6,DAT_1803b2de0)
                                        ;
                                        if ((-1 < (int)*plVar6) &&
                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                          lVar16 = FUN_18000cc40(plVar6);
                                          (**(code **)(lVar16 + 0x30))(plVar6);
                                        }
                                        if (pcVar9 == (code *)0x0) {
                                          FUN_18000ce30(param_1,&local_2d8);
                                          plVar6 = local_2d0;
                                          plVar8 = local_2b8;
                                        }
                                        else {
                                          pcVar10 = pcVar9;
                                          local_218 = pcVar9;
                                          if (pcVar9 == _Py_NoneStruct_exref) {
                                            plVar8 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
                                                     FUN_18037f0c0(param_1,plVar8,DAT_1803b2b00);
                                            if (plVar8 == (longlong *)0x0) {
                                              FUN_18000ce30(param_1,&local_2d8);
                                              uVar17 = 0xe3;
                                              plVar6 = local_2d0;
                                              plVar8 = local_2b8;
                                            }
                                            else {
                                              uVar17 = 0xe3;
                                              *(undefined4 *)(plVar15 + 5) = 0xe3;
                                              pcVar10 = (code *)FUN_1803770d0(param_1,plVar8,
                                                                              DAT_1803b2de8);
                                              if ((-1 < (int)*plVar8) &&
                                                 (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                lVar16 = FUN_18000cc40(plVar8);
                                                (**(code **)(lVar16 + 0x30))(plVar8);
                                              }
                                              if (pcVar10 != (code *)0x0) {
                                                local_218 = pcVar10;
                                                if ((-1 < *(int *)pcVar9) &&
                                                   (*(longlong *)pcVar9 = *(longlong *)pcVar9 + -1,
                                                   *(longlong *)pcVar9 == 0)) {
                                                  lVar16 = FUN_18000cc40(pcVar9);
                                                  (**(code **)(lVar16 + 0x30))(pcVar9);
                                                }
                                                goto LAB_1800ce03a;
                                              }
                                              FUN_18000ce30(param_1,&local_2d8);
                                              plVar6 = local_2d0;
                                              plVar8 = local_2b8;
                                            }
                                          }
                                          else {
LAB_1800ce03a:
                                            if (pcVar10 == _Py_NoneStruct_exref) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#94="max" bytes=[3953977, 3953982] */
                                              uVar12 = FUN_18000d570(DAT_1803b2df0);
                                              uVar2 = DAT_1803b2e08;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#95="min" bytes=[3953982, 3953987] */
                                              uVar13 = FUN_18000d570(DAT_1803b2df8);
                                              uVar11 = DAT_1803b2e10;
                                              plVar8 = (longlong *)
                                                       FUN_1803813f0(local_268,DAT_1803b2e18);
                                              if (plVar8 == (longlong *)0x0) {
                                                FUN_18000ce30(param_1,&local_2d8);
                                                uVar17 = 0xe7;
                                                plVar15 = local_2b0;
                                                plVar5 = local_2c8;
                                                plVar6 = local_2d0;
                                                plVar8 = local_2b8;
                                              }
                                              else {
                                                uVar17 = 0xe7;
                                                local_198 = uVar11;
                                                *(undefined4 *)(local_2b0 + 5) = 0xe7;
                                                local_190 = plVar8;
                                                plVar6 = (longlong *)
                                                         FUN_180377750(param_1,uVar13,&local_198);
                                                if ((-1 < (int)*plVar8) &&
                                                   (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                  lVar16 = FUN_18000cc40(plVar8);
                                                  (**(code **)(lVar16 + 0x30))(plVar8);
                                                }
                                                if (plVar6 == (longlong *)0x0) {
                                                  FUN_18000ce30(param_1,&local_2d8);
                                                  plVar15 = local_2b0;
                                                  plVar5 = local_2c8;
                                                  plVar6 = local_2d0;
                                                  plVar8 = local_2b8;
                                                }
                                                else {
                                                  local_120 = uVar2;
                                                  *(undefined4 *)(local_2b0 + 5) = 0xe7;
                                                  local_118 = plVar6;
                                                  plVar8 = (longlong *)
                                                           FUN_180377750(param_1,uVar12,&local_120);
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar6);
                                                    (**(code **)(lVar16 + 0x30))(plVar6);
                                                  }
                                                  if (plVar8 != (longlong *)0x0) goto LAB_1800ce26f;
                                                  FUN_18000ce30(param_1,&local_2d8);
                                                  plVar15 = local_2b0;
                                                  plVar5 = local_2c8;
                                                  plVar6 = local_2d0;
                                                  plVar8 = local_2b8;
                                                }
                                              }
                                            }
                                            else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#94="max" bytes=[3953977, 3953982] */
                                              uVar11 = FUN_18000d570(DAT_1803b2df0);
                                              uVar2 = DAT_1803ab258;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#95="min" bytes=[3953982, 3953987] */
                                              uVar12 = FUN_18000d570(DAT_1803b2df8);
                                              plVar8 = DAT_1803b2e00;
                                              plVar6 = (longlong *)PyNumber_Long(pcVar10);
                                              if (plVar6 == (longlong *)0x0) {
                                                FUN_18000ce30(param_1,&local_2d8);
                                                uVar17 = 0xe5;
                                                plVar15 = local_2b0;
                                                plVar5 = local_2c8;
                                                plVar6 = local_2d0;
                                                plVar8 = local_2b8;
                                              }
                                              else {
                                                uVar17 = 0xe5;
                                                local_288 = plVar8;
                                                *(undefined4 *)(local_2b0 + 5) = 0xe5;
                                                local_280 = plVar6;
                                                plVar5 = (longlong *)
                                                         FUN_180377750(param_1,uVar12,&local_288);
                                                if ((-1 < (int)*plVar6) &&
                                                   (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                  lVar16 = FUN_18000cc40(plVar6);
                                                  (**(code **)(lVar16 + 0x30))(plVar6);
                                                }
                                                if (plVar5 == (longlong *)0x0) {
                                                  FUN_18000ce30(param_1,&local_2d8);
                                                  plVar15 = local_2b0;
                                                  plVar5 = local_2c8;
                                                  plVar6 = local_2d0;
                                                  plVar8 = local_2b8;
                                                }
                                                else {
                                                  local_198 = uVar2;
                                                  *(undefined4 *)(local_2b0 + 5) = 0xe5;
                                                  local_190 = plVar5;
                                                  plVar8 = (longlong *)
                                                           FUN_180377750(param_1,uVar11,&local_198);
                                                  if ((-1 < (int)*plVar5) &&
                                                     (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar5);
                                                    (**(code **)(lVar16 + 0x30))(plVar5);
                                                  }
                                                  if (plVar8 == (longlong *)0x0) goto LAB_1800ce146;
LAB_1800ce26f:
                                                  local_258 = plVar8;
                                                  iVar3 = FUN_1803841f0(plVar8,DAT_1803ab258);
                                                  if (iVar3 == -1) {
                                                    FUN_18000ce30(param_1,&local_2d8);
                                                    uVar17 = 0xe8;
                                                    plVar15 = local_2b0;
                                                    plVar5 = local_2c8;
                                                    plVar6 = local_2d0;
                                                    plVar8 = local_2b8;
                                                  }
                                                  else if (iVar3 == 1) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#94="max" bytes=[3953977, 3953982] */
                                                    uVar11 = FUN_18000d570(DAT_1803b2df0);
                                                    uVar2 = DAT_1803b2e18;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#95="min" bytes=[3953982, 3953987] */
                                                    uVar12 = FUN_18000d570(DAT_1803b2df8);
                                                    plVar4 = local_2b0;
                                                    uVar17 = 0xe9;
                                                    local_288 = DAT_1803b2e20;
                                                    local_280 = local_268;
                                                    *(undefined4 *)(local_2b0 + 5) = 0xe9;
                                                    plVar8 = (longlong *)
                                                             FUN_180377750(param_1,uVar12,&local_288
                                                                          );
                                                    plVar15 = plVar4;
                                                    if (plVar8 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_2d8);
                                                      plVar5 = local_2c8;
                                                      plVar6 = local_2d0;
                                                      plVar8 = local_2b8;
                                                    }
                                                    else {
                                                      *(undefined4 *)(plVar4 + 5) = 0xe9;
                                                      local_180 = uVar2;
                                                      local_178 = plVar8;
                                                      plVar5 = (longlong *)
                                                               FUN_180377750(param_1,uVar11,
                                                                             &local_180);
                                                      if ((-1 < (int)*plVar8) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar8);
                                                        (**(code **)(lVar16 + 0x30))(plVar8);
                                                      }
                                                      if (plVar5 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_2d8);
                                                        plVar5 = local_2c8;
                                                        plVar6 = local_2d0;
                                                        plVar8 = local_2b8;
                                                      }
                                                      else {
LAB_1800ce481:
                                                        plVar6 = local_2d0;
                                                        local_250 = plVar5;
                                                        iVar3 = FUN_180001510(local_2d0);
                                                        plVar8 = local_1e8;
                                                        plVar15 = plVar4;
                                                        if (iVar3 == -1) {
                                                          FUN_18000ce30(param_1,&local_2d8);
                                                          uVar17 = 0xed;
                                                          plVar5 = local_2c8;
                                                          plVar8 = local_2b8;
                                                        }
                                                        else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#102="基于拆书报告的仿写创作" bytes=[3954003, 3954038] */
                                                          plVar5 = DAT_1803b2e30;
                                                          if (iVar3 == 1) {
                                                            plVar5 = plVar6;
                                                          }
                                                          iVar3 = (int)*plVar5 + 1;
                                                          if (iVar3 != 0) {
                                                            *(int *)plVar5 = iVar3;
                                                          }
                                                          local_248 = plVar5;
                                                          iVar3 = FUN_180001510(local_1e8);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#103="[" bytes=[3954038, 3954040] */
                                                          piVar1 = DAT_1803b2e38;
                                                          if (iVar3 == -1) {
                                                            FUN_18000ce30(param_1,&local_2d8);
                                                            uVar17 = 0xee;
                                                            plVar5 = local_2c8;
                                                            plVar8 = local_2b8;
                                                          }
                                                          else {
                                                            plVar7 = plVar5;
                                                            if (iVar3 == 0) {
LAB_1800ce658:
                                                              plVar8 = local_220;
                                                              iVar3 = FUN_180001510(local_220);
                                                              if (iVar3 == -1) {
                                                                FUN_18000ce30(param_1,&local_2d8);
                                                                uVar17 = 0xf0;
                                                                plVar5 = local_2c8;
                                                                plVar8 = local_2b8;
                                                              }
                                                              else if (iVar3 == 0) {
LAB_1800ce7a8:
                                                                lVar16 = FUN_1800c7500(param_1);
                                                                if (lVar16 == 0) {
                                                                  FUN_180373340(param_1,&local_2d8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#106="build_imitation_short_writer_addendum" bytes=[3954063, 3954102] */
                                                                                DAT_1803b2e50);
                                                                  uVar17 = 0xf3;
                                                                  plVar15 = local_2b0;
                                                                  plVar5 = local_2c8;
                                                                  plVar6 = local_2d0;
                                                                  plVar8 = local_2b8;
                                                                }
                                                                else {
                                                                  plVar6 = (longlong *)
                                                                           FUN_180024e50(param_1,
                                                  local_1a0,DAT_1803b2e58);
                                                  plVar8 = local_1d0;
                                                  if (plVar6 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_2d8);
                                                    uVar17 = 0xf4;
                                                    plVar15 = local_2b0;
                                                    plVar5 = local_2c8;
                                                    plVar6 = local_2d0;
                                                    plVar8 = local_2b8;
                                                  }
                                                  else {
                                                    plVar4 = (longlong *)
                                                             FUN_180024e50(param_1,local_1d0,
                                                                           DAT_1803b2e60);
                                                    plVar15 = local_2b0;
                                                    if (plVar4 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_2d8);
                                                      if ((-1 < (int)*plVar6) &&
                                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar6);
                                                        (**(code **)(lVar16 + 0x30))(plVar6);
                                                      }
                                                      uVar17 = 0xf5;
                                                      plVar15 = local_2b0;
                                                      plVar5 = local_2c8;
                                                      plVar6 = local_2d0;
                                                      plVar8 = local_2b8;
                                                    }
                                                    else {
                                                      uVar17 = 0xf6;
                                                      *(undefined4 *)(local_2b0 + 5) = 0xf6;
                                                      plVar8 = (longlong *)
                                                               FUN_18037e6c0(param_1,local_2b8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#109="_extract_anti_plagiarism" bytes=[3954114, 3954140] */
                                                                             DAT_1803b2e68,plVar8);
                                                      if (plVar8 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_2d8);
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          lVar16 = FUN_18000cc40(plVar6);
                                                          (**(code **)(lVar16 + 0x30))(plVar6);
                                                        }
                                                        plVar5 = local_2c8;
                                                        plVar6 = local_2d0;
                                                        plVar8 = local_2b8;
                                                        if ((-1 < (int)*plVar4) &&
                                                           (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                                                          lVar16 = FUN_18000cc40(plVar4);
                                                          (**(code **)(lVar16 + 0x30))(plVar4);
                                                          plVar5 = local_2c8;
                                                          plVar6 = local_2d0;
                                                          plVar8 = local_2b8;
                                                        }
                                                      }
                                                      else {
                                                        uVar17 = 0xf3;
                                                        *(undefined4 *)(plVar15 + 5) = 0xf3;
                                                        local_110 = plVar6;
                                                        local_108 = plVar4;
                                                        local_100 = plVar8;
                                                        plVar5 = (longlong *)
                                                                 FUN_18037ccb0(param_1,lVar16,
                                                                               &local_110,
                                                                               DAT_1803b2e70);
                                                        local_2c0 = plVar5;
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          lVar16 = FUN_18000cc40(plVar6);
                                                          (**(code **)(lVar16 + 0x30))(plVar6);
                                                        }
                                                        if ((-1 < (int)*plVar4) &&
                                                           (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                                                          lVar16 = FUN_18000cc40(plVar4);
                                                          (**(code **)(lVar16 + 0x30))(plVar4);
                                                        }
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar16 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar16 + 0x30))(plVar8);
                                                        }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#111="direction" bytes=[3954194, 3954205] */
                                                        uVar2 = DAT_1803b2e78;
                                                        if (plVar5 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_2d8);
                                                          plVar5 = local_2c8;
                                                          plVar6 = local_2d0;
                                                          plVar8 = local_2b8;
                                                        }
                                                        else {
                                                          local_1c0 = plVar5;
                                                          plVar4 = (longlong *)
                                                                   _PyDict_NewPresized(6);
                                                          local_270 = plVar4;
                                                          PyDict_SetItem(plVar4,uVar2,local_248);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#112="chapters" bytes=[3954205, 3954215] */
                                                          PyDict_SetItem(plVar4,DAT_1803b2e80,
                                                                         local_258);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#113="charsPerChapter" bytes=[3954215, 3954232] */
                                                          PyDict_SetItem(plVar4,DAT_1803b2e88,
                                                                         local_250);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#114="referenceText" bytes=[3954232, 3954247] */
                                                          uVar17 = DAT_1803b2e90;
                                                          plVar8 = (longlong *)
                                                                   FUN_180024e50(param_1,local_230,
                                                                                 DAT_1803b2e60);
                                                          if (plVar8 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_2d8);
                                                            uVar17 = 0xfd;
                                                          }
                                                          else {
                                                            PyDict_SetItem(plVar4,uVar17,plVar8);
                                                            if ((-1 < (int)*plVar8) &&
                                                               (*plVar8 = *plVar8 + -1, *plVar8 == 0
                                                               )) {
                                                              lVar16 = FUN_18000cc40(plVar8);
                                                              (**(code **)(lVar16 + 0x30))(plVar8);
                                                            }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#65="storyId" bytes=[3953414, 3953423] */
                                                            PyDict_SetItem(plVar4,DAT_1803b2d08,
                                                                           local_278);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#66="title" bytes=[3953423, 3953430] */
                                                            plVar8 = DAT_1803b2d10;
                                                            plVar6 = (longlong *)
                                                                     FUN_18037f0c0(param_1,local_260
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
                                                                                   ,DAT_1803b2b00);
                                                            if (plVar6 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_2d8);
                                                              uVar17 = 0xff;
                                                            }
                                                            else {
                                                              uVar17 = 0xff;
                                                              *(undefined4 *)(local_2b0 + 5) = 0xff;
                                                              plVar15 = (longlong *)
                                                                        FUN_1803770d0(param_1,plVar6
                                                                                      ,DAT_1803b2b40
                                                                                     );
                                                              if ((-1 < (int)*plVar6) &&
                                                                 (*plVar6 = *plVar6 + -1,
                                                                 *plVar6 == 0)) {
                                                                lVar16 = FUN_18000cc40(plVar6);
                                                                (**(code **)(lVar16 + 0x30))(plVar6)
                                                                ;
                                                              }
                                                              if (plVar15 == (longlong *)0x0) {
                                                                FUN_18000ce30(param_1,&local_2d8);
                                                              }
                                                              else {
                                                                iVar3 = FUN_180001510(plVar15);
                                                                if (iVar3 != -1) {
                                                                  if (iVar3 != 1) {
                                                                    if ((-1 < (int)*plVar15) &&
                                                                       (*plVar15 = *plVar15 + -1,
                                                                       *plVar15 == 0)) {
                                                                      lVar16 = FUN_18000cc40(plVar15
                                                  );
                                                  (**(code **)(lVar16 + 0x30))(plVar15);
                                                  }
                                                  iVar3 = (int)*local_278 + 1;
                                                  plVar15 = local_278;
                                                  if (iVar3 != 0) {
                                                    *(int *)local_278 = iVar3;
                                                  }
                                                  }
                                                  PyDict_SetItem(plVar4,plVar8,plVar15);
                                                  if ((-1 < (int)*plVar15) &&
                                                     (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar15);
                                                    (**(code **)(lVar16 + 0x30))(plVar15);
                                                  }
                                                  plVar15 = local_2b0;
                                                  plVar8 = local_2b8;
                                                  uVar17 = 0x102;
                                                  local_f8 = local_2c8;
                                                  local_f0 = local_278;
                                                  *(undefined4 *)(local_2b0 + 5) = 0x102;
                                                  local_e8 = local_260;
                                                  local_1b8 = plVar4;
                                                  local_e0 = plVar4;
                                                  plVar7 = (longlong *)
                                                           FUN_18037eae0(param_1,local_2b8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#115="_maybe_pause_after_blueprint" bytes=[3954247, 3954277] */
                                                                         DAT_1803b2e98,&local_f8);
                                                  local_288 = plVar7;
                                                  if (plVar7 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_2d8);
                                                    plVar5 = local_2c8;
                                                    plVar6 = local_2d0;
                                                  }
                                                  else {
                                                    local_1b0 = plVar7;
                                                    iVar3 = FUN_180001510(plVar7);
                                                    if (iVar3 == -1) {
                                                      FUN_18000ce30(param_1,&local_2d8);
                                                      uVar17 = 0x103;
                                                      plVar5 = local_2c8;
                                                      plVar6 = local_2d0;
                                                    }
                                                    else {
                                                      if (iVar3 != 0) {
                                                        iVar3 = (int)*plVar7 + 1;
                                                        plVar6 = plVar7;
                                                        if (iVar3 != 0) {
                                                          *(int *)plVar7 = iVar3;
                                                        }
                                                        goto LAB_1800d03d5;
                                                      }
                                                      local_290 = FUN_1800c7350(param_1);
                                                      if (local_290 == 0) {
                                                        FUN_180373340(param_1,&local_2d8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#116="ShortFictionRunner" bytes=[3954277, 3954297] */
                                                                      DAT_1803b2ea0);
                                                        uVar17 = 0x106;
                                                        plVar5 = local_2c8;
                                                        plVar6 = local_2d0;
                                                      }
                                                      else {
                                                        plVar4 = (longlong *)
                                                                 FUN_18037f0c0(param_1,plVar8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#14="shorts_root" bytes=[3952645, 3952658] */
                                                                               DAT_1803b2b70);
                                                        if (plVar4 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_2d8);
                                                          uVar17 = 0x107;
                                                          plVar5 = local_2c8;
                                                          plVar6 = local_2d0;
                                                        }
                                                        else {
                                                          uVar17 = 0x108;
                                                          *(undefined4 *)(plVar15 + 5) = 0x108;
                                                          plVar7 = (longlong *)
                                                                   FUN_18037e6c0(param_1,plVar8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#117="_make_augmented_ctx" bytes=[3954297, 3954318] */
                                                                                 DAT_1803b2ea8,
                                                                                 plVar5);
                                                          plVar15 = local_2b0;
                                                          if (plVar7 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_2d8);
                                                            plVar15 = local_2b0;
                                                            plVar5 = local_2c8;
                                                            plVar6 = local_2d0;
                                                            if ((-1 < (int)*plVar4) &&
                                                               (*plVar4 = *plVar4 + -1, *plVar4 == 0
                                                               )) {
                                                              lVar16 = FUN_18000cc40(plVar4);
                                                              (**(code **)(lVar16 + 0x30))(plVar4);
                                                              plVar15 = local_2b0;
                                                              plVar5 = local_2c8;
                                                              plVar6 = local_2d0;
                                                            }
                                                          }
                                                          else {
                                                            uVar17 = 0x109;
                                                            *(undefined4 *)(local_2b0 + 5) = 0x109;
                                                            plVar6 = (longlong *)
                                                                     FUN_18037e6c0(param_1,plVar8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#118="_make_offset_emitter" bytes=[3954318, 3954340] */
                                                                                   DAT_1803b2eb0,
                                                                                   *(undefined8 *)
                                                                                    (DAT_1803b2eb8 +
                                                                                    0x18));
                                                            plVar8 = local_2b8;
                                                            if (plVar6 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_2d8);
                                                              if ((-1 < (int)*plVar4) &&
                                                                 (*plVar4 = *plVar4 + -1,
                                                                 *plVar4 == 0)) {
                                                                lVar16 = FUN_18000cc40(plVar4);
                                                                (**(code **)(lVar16 + 0x30))(plVar4)
                                                                ;
                                                              }
                                                              plVar5 = local_2c8;
                                                              plVar6 = local_2d0;
                                                              plVar8 = local_2b8;
                                                              if ((-1 < (int)*plVar7) &&
                                                                 (*plVar7 = *plVar7 + -1,
                                                                 *plVar7 == 0)) {
                                                                lVar16 = FUN_18000cc40(plVar7);
                                                                (**(code **)(lVar16 + 0x30))(plVar7)
                                                                ;
                                                                plVar5 = local_2c8;
                                                                plVar6 = local_2d0;
                                                                plVar8 = local_2b8;
                                                              }
                                                            }
                                                            else {
                                                              plVar5 = (longlong *)
                                                                       FUN_18037f0c0(param_1,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#19="_cancelled" bytes=[3952739, 3952751] */
                                                  local_2b8,DAT_1803b2b98);
                                                  if (plVar5 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_2d8);
                                                    if ((-1 < (int)*plVar4) &&
                                                       (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                                                      lVar16 = FUN_18000cc40(plVar4);
                                                      (**(code **)(lVar16 + 0x30))(plVar4);
                                                    }
                                                    if ((-1 < (int)*plVar7) &&
                                                       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                      lVar16 = FUN_18000cc40(plVar7);
                                                      (**(code **)(lVar16 + 0x30))(plVar7);
                                                    }
                                                    if ((-1 < (int)*plVar6) &&
                                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                      lVar16 = FUN_18000cc40(plVar6);
                                                      (**(code **)(lVar16 + 0x30))(plVar6);
                                                    }
                                                    uVar17 = 0x10a;
                                                    plVar15 = local_2b0;
                                                    plVar5 = local_2c8;
                                                    plVar6 = local_2d0;
                                                    plVar8 = local_2b8;
                                                  }
                                                  else {
                                                    plVar8 = (longlong *)
                                                             FUN_18037f0c0(param_1,plVar8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#21="_review_config" bytes=[3952780, 3952796] */
                                                                           DAT_1803b2ba8);
                                                    if (plVar8 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_2d8);
                                                      if ((-1 < (int)*plVar4) &&
                                                         (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar4);
                                                        (**(code **)(lVar16 + 0x30))(plVar4);
                                                      }
                                                      if ((-1 < (int)*plVar7) &&
                                                         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar7);
                                                        (**(code **)(lVar16 + 0x30))(plVar7);
                                                      }
                                                      if ((-1 < (int)*plVar6) &&
                                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar6);
                                                        (**(code **)(lVar16 + 0x30))(plVar6);
                                                      }
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar5);
                                                        (**(code **)(lVar16 + 0x30))(plVar5);
                                                      }
                                                      uVar17 = 0x10b;
                                                      plVar15 = local_2b0;
                                                      plVar5 = local_2c8;
                                                      plVar6 = local_2d0;
                                                      plVar8 = local_2b8;
                                                    }
                                                    else {
                                                      uVar17 = 0x106;
                                                      local_c0 = _Py_TrueStruct_exref;
                                                      *(undefined4 *)(local_2b0 + 5) = 0x106;
                                                      local_398 = (longlong *)DAT_1803b2ec0;
                                                      local_150 = plVar4;
                                                      local_148 = plVar7;
                                                      local_d8 = plVar6;
                                                      local_d0 = plVar5;
                                                      local_c8 = plVar8;
                                                      local_2a0 = (longlong *)
                                                                  FUN_18037d650(param_1,local_290,
                                                                                &local_150,&local_d8
                                                                               );
                                                      if ((-1 < (int)*plVar4) &&
                                                         (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar4);
                                                        (**(code **)(lVar16 + 0x30))(plVar4);
                                                      }
                                                      if ((-1 < (int)*plVar7) &&
                                                         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar7);
                                                        (**(code **)(lVar16 + 0x30))(plVar7);
                                                      }
                                                      if ((-1 < (int)*plVar6) &&
                                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar6);
                                                        (**(code **)(lVar16 + 0x30))(plVar6);
                                                      }
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar5);
                                                        (**(code **)(lVar16 + 0x30))(plVar5);
                                                      }
                                                      if ((-1 < (int)*plVar8) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar8);
                                                        (**(code **)(lVar16 + 0x30))(plVar8);
                                                      }
                                                      plVar4 = local_2b0;
                                                      if (local_2a0 == (longlong *)0x0)
                                                      goto LAB_1800ce146;
                                                      uVar17 = 0x10e;
                                                      local_1a8 = local_2a0;
                                                      *(undefined4 *)(local_2b0 + 5) = 0x10e;
                                                      plVar6 = (longlong *)
                                                               FUN_18037e6c0(param_1,local_2a0,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#121="run" bytes=[3954400, 3954405] */
                                                                             DAT_1803b2ec8,local_270
                                                                            );
                                                      plVar8 = local_2b8;
                                                      plVar15 = plVar4;
                                                      local_2a8 = plVar6;
                                                      if (plVar6 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_2d8);
                                                        plVar5 = local_2c8;
                                                        plVar6 = local_2d0;
                                                        plVar8 = local_2b8;
                                                      }
                                                      else {
                                                        uVar17 = 0x110;
                                                        *(undefined4 *)(plVar4 + 5) = 0x110;
                                                        local_1f8 = plVar6;
                                                        plVar5 = (longlong *)
                                                                 FUN_18037e560(param_1,local_2b8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#81="_check_cancel" bytes=[3953713, 3953728] */
                                                                               DAT_1803b2d88);
                                                        if (plVar5 == (longlong *)0x0) {
LAB_1800cf062:
                                                          FUN_18000ce30(param_1,&local_2d8);
                                                          plVar15 = plVar4;
                                                          plVar5 = local_2c8;
                                                          plVar6 = local_2d0;
                                                        }
                                                        else {
                                                          if ((-1 < (int)*plVar5) &&
                                                             (*plVar5 = *plVar5 + -1, *plVar5 == 0))
                                                          {
                                                            lVar16 = FUN_18000cc40(plVar5);
                                                            (**(code **)(lVar16 + 0x30))(plVar5);
                                                          }
                                                          plVar5 = (longlong *)
                                                                   FUN_18037f0c0(param_1,plVar6,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
                                                                                 DAT_1803b2b00);
                                                          if (plVar5 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_2d8);
                                                            uVar17 = 0x115;
                                                            plVar5 = local_2c8;
                                                            plVar6 = local_2d0;
                                                          }
                                                          else {
                                                            uVar17 = 0x115;
                                                            *(undefined4 *)(plVar4 + 5) = 0x115;
                                                            plVar7 = (longlong *)
                                                                     FUN_180378210(param_1,plVar5,
                                                                                   DAT_1803b2ed0);
                                                            local_298 = plVar7;
                                                            if ((-1 < (int)*plVar5) &&
                                                               (*plVar5 = *plVar5 + -1, *plVar5 == 0
                                                               )) {
                                                              lVar16 = FUN_18000cc40(plVar5);
                                                              (**(code **)(lVar16 + 0x30))(plVar5);
                                                            }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                            plVar5 = DAT_1803b2b10;
                                                            if (plVar7 == (longlong *)0x0)
                                                            goto LAB_1800cf062;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                            iVar3 = (int)*DAT_1803b2b10 + 1;
                                                            if (iVar3 != 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                              *(int *)DAT_1803b2b10 = iVar3;
                                                            }
                                                            local_240 = plVar5;
                                                            local_1f0 = plVar7;
                                                            iVar3 = FUN_180001510(plVar7);
                                                            if (iVar3 == -1) {
                                                              FUN_18000ce30(param_1,&local_2d8);
                                                              uVar17 = 0x117;
                                                              plVar5 = local_2c8;
                                                              plVar6 = local_2d0;
                                                              plVar8 = local_2b8;
                                                            }
                                                            else {
                                                              plVar8 = plVar5;
                                                              if (iVar3 == 0) {
LAB_1800cf2d2:
                                                                iVar3 = FUN_180001510(plVar8);
                                                                plVar8 = local_2b8;
                                                                plVar15 = plVar4;
                                                                if (iVar3 != -1) {
                                                                  if (iVar3 == 0) {
                                                                    uVar17 = 0x11f;
                                                                    *(undefined4 *)(plVar4 + 5) =
                                                                         0x11f;
                                                                    plVar5 = (longlong *)
                                                                             FUN_18037e980(param_1,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#124="_stage" bytes=[3954433, 3954441] */
                                                  local_2b8,DAT_1803b2ee0,DAT_1803b2f00 + 0x18);
                                                  if (plVar5 != (longlong *)0x0) {
                                                    if ((-1 < (int)*plVar5) &&
                                                       (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                      lVar16 = FUN_18000cc40(plVar5);
                                                      (**(code **)(lVar16 + 0x30))(plVar5);
                                                    }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                    local_238 = DAT_1803b2b10;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                    iVar3 = (int)*DAT_1803b2b10 + 1;
                                                    if (iVar3 != 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                      *(int *)DAT_1803b2b10 = iVar3;
                                                    }
LAB_1800cf447:
                                                    uVar17 = 0x125;
                                                    *(undefined4 *)(plVar4 + 5) = 0x125;
                                                    plVar5 = (longlong *)
                                                             FUN_18037e980(param_1,plVar8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#124="_stage" bytes=[3954433, 3954441] */
                                                                           DAT_1803b2ee0,
                                                                           DAT_1803b2f08 + 0x18);
                                                    if (plVar5 != (longlong *)0x0) {
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar5);
                                                        (**(code **)(lVar16 + 0x30))(plVar5);
                                                      }
                                                      lVar16 = FUN_1800c7020(param_1);
                                                      if (lVar16 == 0) {
                                                        FUN_180373340(param_1,&local_2d8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#4="FileRepository" bytes=[3952488, 3952504] */
                                                                      DAT_1803b2b20);
                                                        uVar17 = 0x127;
                                                        plVar5 = local_2c8;
                                                        plVar6 = local_2d0;
                                                      }
                                                      else {
                                                        plVar5 = (longlong *)
                                                                 FUN_18037f0c0(param_1,lVar16,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#78="write_json" bytes=[3953686, 3953698] */
                                                                               DAT_1803b2d70);
                                                        if (plVar5 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_2d8);
                                                          uVar17 = 0x127;
                                                          plVar5 = local_2c8;
                                                          plVar6 = local_2d0;
                                                        }
                                                        else {
                                                          plVar15 = (longlong *)
                                                                    FUN_180382540(local_2c8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#58="status.json" bytes=[3953343, 3953356] */
                                                                                  DAT_1803b2cd0);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#72="status" bytes=[3953504, 3953512] */
                                                          uVar2 = DAT_1803b2d40;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#60="completed" bytes=[3953366, 3953377] */
                                                          uVar17 = DAT_1803b2ce0;
                                                          if (plVar15 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_2d8);
                                                            if ((-1 < (int)*plVar5) &&
                                                               (*plVar5 = *plVar5 + -1, *plVar5 == 0
                                                               )) {
                                                              lVar16 = FUN_18000cc40(plVar5);
                                                              (**(code **)(lVar16 + 0x30))(plVar5);
                                                            }
                                                            uVar17 = 0x127;
                                                            plVar15 = local_2b0;
                                                            plVar5 = local_2c8;
                                                            plVar6 = local_2d0;
                                                          }
                                                          else {
                                                            plVar4 = (longlong *)
                                                                     _PyDict_NewPresized(5);
                                                            PyDict_SetItem(plVar4,uVar2,uVar17);
                                                            plVar8 = local_278;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#65="storyId" bytes=[3953414, 3953423] */
                                                            PyDict_SetItem(plVar4,DAT_1803b2d08,
                                                                           local_278);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#66="title" bytes=[3953423, 3953430] */
                                                            local_2a0 = DAT_1803b2d10;
                                                            plVar6 = (longlong *)
                                                                     FUN_18037f0c0(param_1,plVar6,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
                                                                                   DAT_1803b2b00);
                                                            uVar17 = 0x12a;
                                                            if (plVar6 != (longlong *)0x0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#66="title" bytes=[3953423, 3953430] */
                                                              local_208 = DAT_1803b2d10;
                                                              local_200 = plVar8;
                                                              *(undefined4 *)(local_2b0 + 5) = 0x12a
                                                              ;
                                                              plVar8 = (longlong *)
                                                                       FUN_180377750(param_1,plVar6,
                                                                                     &local_208);
                                                              if ((-1 < (int)*plVar6) &&
                                                                 (*plVar6 = *plVar6 + -1,
                                                                 *plVar6 == 0)) {
                                                                lVar16 = FUN_18000cc40(plVar6);
                                                                (**(code **)(lVar16 + 0x30))(plVar6)
                                                                ;
                                                              }
                                                              if (plVar8 != (longlong *)0x0) {
                                                                PyDict_SetItem(plVar4,local_2a0,
                                                                               plVar8);
                                                                if ((-1 < (int)*plVar8) &&
                                                                   (*plVar8 = *plVar8 + -1,
                                                                   *plVar8 == 0)) {
                                                                  lVar16 = FUN_18000cc40(plVar8);
                                                                  (**(code **)(lVar16 + 0x30))
                                                                            (plVar8);
                                                                }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#130="wordCount" bytes=[3954569, 3954580] */
                                                                local_2a0 = DAT_1803b2f10;
                                                                plVar8 = (longlong *)
                                                                         FUN_18037f0c0(param_1,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
                                                  local_2a8,DAT_1803b2b00);
                                                  uVar17 = 299;
                                                  if (plVar8 != (longlong *)0x0) {
                                                    *(undefined4 *)(local_2b0 + 5) = 299;
                                                    plVar6 = (longlong *)
                                                             FUN_180378210(param_1,plVar8,
                                                                           DAT_1803b2f18);
                                                    if ((-1 < (int)*plVar8) &&
                                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                      lVar16 = FUN_18000cc40(plVar8);
                                                      (**(code **)(lVar16 + 0x30))(plVar8);
                                                    }
                                                    if (plVar6 != (longlong *)0x0) {
                                                      PyDict_SetItem(plVar4,local_2a0,plVar6);
                                                      if ((-1 < (int)*plVar6) &&
                                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar6);
                                                        (**(code **)(lVar16 + 0x30))(plVar6);
                                                      }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#132="mode" bytes=[3954595, 3954601] */
                                                      PyDict_SetItem(plVar4,DAT_1803b2f20,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#133="imitation_short" bytes=[3954601, 3954618] */
                                                                     DAT_1803b2f28);
                                                      uVar17 = 0x127;
                                                      *(undefined4 *)(local_2b0 + 5) = 0x127;
                                                      local_208 = plVar15;
                                                      local_200 = plVar4;
                                                      plVar8 = (longlong *)
                                                               FUN_180377750(param_1,plVar5,
                                                                             &local_208);
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar5);
                                                        (**(code **)(lVar16 + 0x30))(plVar5);
                                                      }
                                                      if ((-1 < (int)*plVar15) &&
                                                         (*plVar15 = *plVar15 + -1, *plVar15 == 0))
                                                      {
                                                        lVar16 = FUN_18000cc40(plVar15);
                                                        (**(code **)(lVar16 + 0x30))(plVar15);
                                                      }
                                                      if ((-1 < (int)*plVar4) &&
                                                         (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar4);
                                                        (**(code **)(lVar16 + 0x30))(plVar4);
                                                      }
                                                      if (plVar8 == (longlong *)0x0) {
LAB_1800ce146:
                                                        FUN_18000ce30(param_1,&local_2d8);
                                                        plVar15 = local_2b0;
                                                        plVar5 = local_2c8;
                                                        plVar6 = local_2d0;
                                                        plVar8 = local_2b8;
                                                      }
                                                      else {
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar16 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar16 + 0x30))(plVar8);
                                                        }
                                                        plVar4 = local_2b0;
                                                        plVar8 = local_2b8;
                                                        uVar17 = 0x12f;
                                                        *(undefined4 *)(local_2b0 + 5) = 0x12f;
                                                        plVar6 = (longlong *)
                                                                 FUN_18037e820(param_1,local_2b8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#124="_stage" bytes=[3954433, 3954441] */
                                                                               DAT_1803b2ee0,
                                                                               DAT_1803b2f30 + 0x18)
                                                        ;
                                                        if (plVar6 == (longlong *)0x0)
                                                        goto LAB_1800cf062;
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          lVar16 = FUN_18000cc40(plVar6);
                                                          (**(code **)(lVar16 + 0x30))(plVar6);
                                                        }
                                                        plVar7 = (longlong *)
                                                                 FUN_18037f0c0(param_1,plVar8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#18="_emit" bytes=[3952732, 3952739] */
                                                                               DAT_1803b2b90);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#69="imitation_short_run" bytes=[3953458, 3953479] */
                                                        uVar2 = DAT_1803b2d28;
                                                        pcVar9 = _Py_TrueStruct_exref;
                                                        if (plVar7 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_2d8);
                                                          uVar17 = 0x130;
                                                          plVar15 = plVar4;
                                                          plVar5 = local_2c8;
                                                          plVar6 = local_2d0;
                                                        }
                                                        else {
                                                          plVar6 = (longlong *)
                                                                   FUN_18037f0c0(param_1,local_2a8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
                                                                                 DAT_1803b2b00);
                                                          if (plVar6 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_2d8);
                                                            if ((-1 < (int)*plVar7) &&
                                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0
                                                               )) {
                                                              lVar16 = FUN_18000cc40(plVar7);
                                                              (**(code **)(lVar16 + 0x30))(plVar7);
                                                            }
                                                            uVar17 = 0x131;
                                                            plVar15 = local_2b0;
                                                            plVar5 = local_2c8;
                                                            plVar6 = local_2d0;
                                                          }
                                                          else {
                                                            uVar17 = 0x131;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#66="title" bytes=[3953423, 3953430] */
                                                            local_208 = DAT_1803b2d10;
                                                            local_200 = local_278;
                                                            *(undefined4 *)(local_2b0 + 5) = 0x131;
                                                            plVar8 = (longlong *)
                                                                     FUN_180377750(param_1,plVar6,
                                                                                   &local_208);
                                                            if ((-1 < (int)*plVar6) &&
                                                               (*plVar6 = *plVar6 + -1, *plVar6 == 0
                                                               )) {
                                                              lVar16 = FUN_18000cc40(plVar6);
                                                              (**(code **)(lVar16 + 0x30))(plVar6);
                                                            }
                                                            if (plVar8 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_2d8);
                                                              plVar15 = local_2b0;
                                                              plVar5 = local_2c8;
                                                              plVar6 = local_2d0;
                                                              plVar8 = local_2b8;
                                                              if ((-1 < (int)*plVar7) &&
                                                                 (*plVar7 = *plVar7 + -1,
                                                                 *plVar7 == 0)) {
                                                                lVar16 = FUN_18000cc40(plVar7);
                                                                (**(code **)(lVar16 + 0x30))(plVar7)
                                                                ;
                                                                plVar15 = local_2b0;
                                                                plVar5 = local_2c8;
                                                                plVar6 = local_2d0;
                                                                plVar8 = local_2b8;
                                                              }
                                                            }
                                                            else {
                                                              uVar17 = 0x130;
                                                              local_b8 = uVar2;
                                                              local_b0 = pcVar9;
                                                              *(undefined4 *)(local_2b0 + 5) = 0x130
                                                              ;
                                                              local_a8 = local_278;
                                                              local_398 = (longlong *)DAT_1803b2f40;
                                                              local_a0 = plVar8;
                                                              plVar6 = (longlong *)
                                                                       FUN_18037d280(param_1,plVar7,
                                                                                     DAT_1803b2f38,
                                                                                     &local_b8);
                                                              if ((-1 < (int)*plVar7) &&
                                                                 (*plVar7 = *plVar7 + -1,
                                                                 *plVar7 == 0)) {
                                                                lVar16 = FUN_18000cc40(plVar7);
                                                                (**(code **)(lVar16 + 0x30))(plVar7)
                                                                ;
                                                              }
                                                              if ((-1 < (int)*plVar8) &&
                                                                 (*plVar8 = *plVar8 + -1,
                                                                 *plVar8 == 0)) {
                                                                lVar16 = FUN_18000cc40(plVar8);
                                                                (**(code **)(lVar16 + 0x30))(plVar8)
                                                                ;
                                                              }
                                                              if (plVar6 == (longlong *)0x0)
                                                              goto LAB_1800ce146;
                                                              if ((-1 < (int)*plVar6) &&
                                                                 (*plVar6 = *plVar6 + -1,
                                                                 *plVar6 == 0)) {
                                                                lVar16 = FUN_18000cc40(plVar6);
                                                                (**(code **)(lVar16 + 0x30))(plVar6)
                                                                ;
                                                              }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#65="storyId" bytes=[3953414, 3953423] */
                                                              uVar17 = DAT_1803b2d08;
                                                              plVar7 = (longlong *)
                                                                       _PyDict_NewPresized(8);
                                                              plVar5 = local_278;
                                                              PyDict_SetItem(plVar7,uVar17,local_278
                                                                            );
                                                              plVar6 = local_2a8;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#66="title" bytes=[3953423, 3953430] */
                                                              plVar8 = DAT_1803b2d10;
                                                              plVar15 = (longlong *)
                                                                        FUN_18037f0c0(param_1,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
                                                  local_2a8,DAT_1803b2b00);
                                                  if (plVar15 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_2d8);
                                                    uVar17 = 0x135;
                                                    plVar5 = local_2c8;
                                                  }
                                                  else {
                                                    uVar17 = 0x135;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#66="title" bytes=[3953423, 3953430] */
                                                    local_208 = DAT_1803b2d10;
                                                    local_200 = plVar5;
                                                    *(undefined4 *)(local_2b0 + 5) = 0x135;
                                                    plVar5 = (longlong *)
                                                             FUN_180377750(param_1,plVar15,
                                                                           &local_208);
                                                    if ((-1 < (int)*plVar15) &&
                                                       (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                                      lVar16 = FUN_18000cc40(plVar15);
                                                      (**(code **)(lVar16 + 0x30))(plVar15);
                                                    }
                                                    if (plVar5 == (longlong *)0x0) {
LAB_1800cfad8:
                                                      FUN_18000ce30(param_1,&local_2d8);
                                                      plVar5 = local_2c8;
                                                    }
                                                    else {
                                                      PyDict_SetItem(plVar7,plVar8,plVar5);
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar5);
                                                        (**(code **)(lVar16 + 0x30))(plVar5);
                                                      }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#67="finalMarkdownPath" bytes=[3953430, 3953449] */
                                                      PyDict_SetItem(plVar7,DAT_1803b2d18,local_298)
                                                      ;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#130="wordCount" bytes=[3954569, 3954580] */
                                                      plVar8 = DAT_1803b2f10;
                                                      plVar6 = (longlong *)
                                                               FUN_18037f0c0(param_1,plVar6,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
                                                                             DAT_1803b2b00);
                                                      if (plVar6 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_2d8);
                                                        uVar17 = 0x137;
                                                        plVar5 = local_2c8;
                                                      }
                                                      else {
                                                        uVar17 = 0x137;
                                                        *(undefined4 *)(local_2b0 + 5) = 0x137;
                                                        plVar5 = (longlong *)
                                                                 FUN_180378210(param_1,plVar6,
                                                                               DAT_1803b2f18);
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          lVar16 = FUN_18000cc40(plVar6);
                                                          (**(code **)(lVar16 + 0x30))(plVar6);
                                                        }
                                                        if (plVar5 == (longlong *)0x0)
                                                        goto LAB_1800cfad8;
                                                        PyDict_SetItem(plVar7,plVar8,plVar5);
                                                        if ((-1 < (int)*plVar5) &&
                                                           (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                          lVar16 = FUN_18000cc40(plVar5);
                                                          (**(code **)(lVar16 + 0x30))(plVar5);
                                                        }
                                                        plVar5 = local_2c8;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#137="deconstructionPath" bytes=[3954678, 3954698] */
                                                        uVar17 = DAT_1803b2f48;
                                                        plVar8 = (longlong *)
                                                                 FUN_180382540(local_2c8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#50="imitation" bytes=[3953271, 3953282] */
                                                                               DAT_1803b2c90);
                                                        if (plVar8 != (longlong *)0x0) {
                                                          plVar6 = (longlong *)
                                                                   FUN_180382540(plVar8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#138="deconstruction.md" bytes=[3954698, 3954717] */
                                                  DAT_1803b2f50);
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar16 + 0x30))(plVar8);
                                                  }
                                                  if (plVar6 != (longlong *)0x0) {
                                                    plVar8 = (longlong *)FUN_1803749a0(plVar6);
                                                    if ((-1 < (int)*plVar6) &&
                                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                      lVar16 = FUN_18000cc40(plVar6);
                                                      (**(code **)(lVar16 + 0x30))(plVar6);
                                                    }
                                                    if (plVar8 != (longlong *)0x0) {
                                                      PyDict_SetItem(plVar7,uVar17,plVar8);
                                                      if ((-1 < (int)*plVar8) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar8);
                                                        (**(code **)(lVar16 + 0x30))(plVar8);
                                                      }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#139="directionsPath" bytes=[3954717, 3954733] */
                                                      uVar17 = DAT_1803b2f58;
                                                      plVar8 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#50="imitation" bytes=[3953271, 3953282] */
                                                               FUN_180382540(plVar5,DAT_1803b2c90);
                                                      if (plVar8 != (longlong *)0x0) {
                                                        plVar6 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#140="directions.md" bytes=[3954733, 3954748] */
                                                                 FUN_180382540(plVar8,DAT_1803b2f60)
                                                        ;
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar16 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar16 + 0x30))(plVar8);
                                                        }
                                                        if (plVar6 != (longlong *)0x0) {
                                                          plVar8 = (longlong *)FUN_1803749a0(plVar6)
                                                          ;
                                                          if ((-1 < (int)*plVar6) &&
                                                             (*plVar6 = *plVar6 + -1, *plVar6 == 0))
                                                          {
                                                            lVar16 = FUN_18000cc40(plVar6);
                                                            (**(code **)(lVar16 + 0x30))(plVar6);
                                                          }
                                                          if (plVar8 != (longlong *)0x0) {
                                                            PyDict_SetItem(plVar7,uVar17,plVar8);
                                                            if ((-1 < (int)*plVar8) &&
                                                               (*plVar8 = *plVar8 + -1, *plVar8 == 0
                                                               )) {
                                                              lVar16 = FUN_18000cc40(plVar8);
                                                              (**(code **)(lVar16 + 0x30))(plVar8);
                                                            }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#141="blueprintPath" bytes=[3954748, 3954763] */
                                                            uVar17 = DAT_1803b2f68;
                                                            plVar8 = (longlong *)
                                                                     FUN_180382540(plVar5,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#50="imitation" bytes=[3953271, 3953282] */
                                                  DAT_1803b2c90);
                                                  if (plVar8 != (longlong *)0x0) {
                                                    plVar6 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#142="blueprint.md" bytes=[3954763, 3954777] */
                                                             FUN_180382540(plVar8,DAT_1803b2f70);
                                                    if ((-1 < (int)*plVar8) &&
                                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                      lVar16 = FUN_18000cc40(plVar8);
                                                      (**(code **)(lVar16 + 0x30))(plVar8);
                                                    }
                                                    if (plVar6 != (longlong *)0x0) {
                                                      plVar8 = (longlong *)FUN_1803749a0(plVar6);
                                                      if ((-1 < (int)*plVar6) &&
                                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar6);
                                                        (**(code **)(lVar16 + 0x30))(plVar6);
                                                      }
                                                      if (plVar8 != (longlong *)0x0) {
                                                        PyDict_SetItem(plVar7,uVar17,plVar8);
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar16 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar16 + 0x30))(plVar8);
                                                        }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#143="editorialReviewPath" bytes=[3954777, 3954798] */
                                                        uVar17 = DAT_1803b2f78;
                                                        iVar3 = FUN_180001510(local_238);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                        plVar15 = DAT_1803b2b10;
                                                        if (iVar3 != -1) {
                                                          if (iVar3 == 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                            iVar3 = (int)*DAT_1803b2b10 + 1;
                                                            if (iVar3 != 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                              *(int *)DAT_1803b2b10 = iVar3;
                                                            }
LAB_1800d039e:
                                                            PyDict_SetItem(plVar7,uVar17,plVar15);
                                                            plVar6 = local_288;
                                                            plVar4 = local_270;
                                                            plVar8 = local_2b8;
                                                            if ((-1 < (int)*plVar15) &&
                                                               (*plVar15 = *plVar15 + -1,
                                                               *plVar15 == 0)) {
                                                              lVar16 = FUN_18000cc40(plVar15);
                                                              (**(code **)(lVar16 + 0x30))(plVar15);
                                                              plVar6 = local_288;
                                                              plVar4 = local_270;
                                                              plVar8 = local_2b8;
                                                            }
LAB_1800d03d5:
                                                            FUN_180001d80(param_1);
                                                            if ((-1 < (int)*local_1a0) &&
                                                               (*local_1a0 = *local_1a0 + -1,
                                                               *local_1a0 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_1e0);
                                                              (**(code **)(lVar16 + 0x30))
                                                                        (local_1e0);
                                                            }
                                                            if ((-1 < (int)*local_188) &&
                                                               (*local_188 = *local_188 + -1,
                                                               *local_188 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_1d8);
                                                              (**(code **)(lVar16 + 0x30))
                                                                        (local_1d8);
                                                            }
                                                            plVar5 = local_2d0;
                                                            if ((-1 < (int)*local_2d0) &&
                                                               (*local_2d0 = *local_2d0 + -1,
                                                               *local_2d0 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_2d0);
                                                              (**(code **)(lVar16 + 0x30))(plVar5);
                                                            }
                                                            if ((-1 < (int)*local_1d0) &&
                                                               (*local_1d0 = *local_1d0 + -1,
                                                               *local_1d0 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_1c8);
                                                              (**(code **)(lVar16 + 0x30))
                                                                        (local_1c8);
                                                            }
                                                            pcVar9 = local_218;
                                                            if ((-1 < *(int *)local_218) &&
                                                               (*(longlong *)local_218 =
                                                                     *(longlong *)local_218 + -1,
                                                               *(longlong *)local_218 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_218);
                                                              (**(code **)(lVar16 + 0x30))(pcVar9);
                                                            }
                                                            plVar5 = local_258;
                                                            if ((-1 < (int)*local_258) &&
                                                               (*local_258 = *local_258 + -1,
                                                               *local_258 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_258);
                                                              (**(code **)(lVar16 + 0x30))(plVar5);
                                                            }
                                                            plVar5 = local_250;
                                                            if ((-1 < (int)*local_250) &&
                                                               (*local_250 = *local_250 + -1,
                                                               *local_250 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_250);
                                                              (**(code **)(lVar16 + 0x30))(plVar5);
                                                            }
                                                            plVar5 = local_248;
                                                            if ((-1 < (int)*local_248) &&
                                                               (*local_248 = *local_248 + -1,
                                                               *local_248 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_248);
                                                              (**(code **)(lVar16 + 0x30))(plVar5);
                                                            }
                                                            if ((-1 < (int)*local_2c0) &&
                                                               (*local_2c0 = *local_2c0 + -1,
                                                               *local_2c0 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_1c0);
                                                              (**(code **)(lVar16 + 0x30))
                                                                        (local_1c0);
                                                            }
                                                            if ((-1 < (int)*plVar4) &&
                                                               (*plVar4 = *plVar4 + -1, *plVar4 == 0
                                                               )) {
                                                              lVar16 = FUN_18000cc40(local_1b8);
                                                              (**(code **)(lVar16 + 0x30))
                                                                        (local_1b8);
                                                            }
                                                            if ((-1 < (int)*plVar6) &&
                                                               (*plVar6 = *plVar6 + -1, *plVar6 == 0
                                                               )) {
                                                              lVar16 = FUN_18000cc40(local_1b0);
                                                              (**(code **)(lVar16 + 0x30))
                                                                        (local_1b0);
                                                            }
                                                            plVar6 = local_1a8;
                                                            if (((local_1a8 != (longlong *)0x0) &&
                                                                (-1 < (int)*local_1a8)) &&
                                                               (*local_1a8 = *local_1a8 + -1,
                                                               *local_1a8 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_1a8);
                                                              (**(code **)(lVar16 + 0x30))(plVar6);
                                                            }
                                                            plVar6 = local_1f8;
                                                            if (((local_1f8 != (longlong *)0x0) &&
                                                                (-1 < (int)*local_1f8)) &&
                                                               (*local_1f8 = *local_1f8 + -1,
                                                               *local_1f8 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_1f8);
                                                              (**(code **)(lVar16 + 0x30))(plVar6);
                                                            }
                                                            plVar6 = local_1f0;
                                                            if (((local_1f0 != (longlong *)0x0) &&
                                                                (-1 < (int)*local_1f0)) &&
                                                               (*local_1f0 = *local_1f0 + -1,
                                                               *local_1f0 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_1f0);
                                                              (**(code **)(lVar16 + 0x30))(plVar6);
                                                            }
                                                            plVar6 = local_240;
                                                            if (((local_240 != (longlong *)0x0) &&
                                                                (-1 < (int)*local_240)) &&
                                                               (*local_240 = *local_240 + -1,
                                                               *local_240 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_240);
                                                              (**(code **)(lVar16 + 0x30))(plVar6);
                                                            }
                                                            plVar6 = local_238;
                                                            if (((local_238 != (longlong *)0x0) &&
                                                                (-1 < (int)*local_238)) &&
                                                               (*local_238 = *local_238 + -1,
                                                               *local_238 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_238);
                                                              (**(code **)(lVar16 + 0x30))(plVar6);
                                                            }
                                                            if ((-1 < (int)*plVar8) &&
                                                               (*plVar8 = *plVar8 + -1, *plVar8 == 0
                                                               )) {
                                                              lVar16 = FUN_18000cc40(plVar8);
                                                              (**(code **)(lVar16 + 0x30))(plVar8);
                                                            }
                                                            plVar8 = local_230;
                                                            if ((-1 < (int)*local_230) &&
                                                               (*local_230 = *local_230 + -1,
                                                               *local_230 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_230);
                                                              (**(code **)(lVar16 + 0x30))(plVar8);
                                                            }
                                                            plVar8 = local_268;
                                                            if ((-1 < (int)*local_268) &&
                                                               (*local_268 = *local_268 + -1,
                                                               *local_268 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_268);
                                                              (**(code **)(lVar16 + 0x30))(plVar8);
                                                            }
                                                            plVar8 = local_228;
                                                            if ((-1 < (int)*local_228) &&
                                                               (*local_228 = *local_228 + -1,
                                                               *local_228 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_228);
                                                              (**(code **)(lVar16 + 0x30))(plVar8);
                                                            }
                                                            plVar8 = local_210;
                                                            if ((-1 < (int)*local_210) &&
                                                               (*local_210 = *local_210 + -1,
                                                               *local_210 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_210);
                                                              (**(code **)(lVar16 + 0x30))(plVar8);
                                                            }
                                                            plVar8 = local_1e8;
                                                            if ((-1 < (int)*local_1e8) &&
                                                               (*local_1e8 = *local_1e8 + -1,
                                                               *local_1e8 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_1e8);
                                                              (**(code **)(lVar16 + 0x30))(plVar8);
                                                            }
                                                            plVar8 = local_220;
                                                            if ((-1 < (int)*local_220) &&
                                                               (*local_220 = *local_220 + -1,
                                                               *local_220 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_220);
                                                              (**(code **)(lVar16 + 0x30))(plVar8);
                                                            }
                                                            plVar8 = local_278;
                                                            if ((-1 < (int)*local_278) &&
                                                               (*local_278 = *local_278 + -1,
                                                               *local_278 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_278);
                                                              (**(code **)(lVar16 + 0x30))(plVar8);
                                                            }
                                                            plVar8 = local_2c8;
                                                            if ((-1 < (int)*local_2c8) &&
                                                               (*local_2c8 = *local_2c8 + -1,
                                                               *local_2c8 == 0)) {
                                                              lVar16 = FUN_18000cc40(local_2c8);
                                                              (**(code **)(lVar16 + 0x30))(plVar8);
                                                            }
                                                            plVar8 = local_260;
                                                            if ((int)*local_260 < 0) {
                                                              return plVar7;
                                                            }
                                                            *local_260 = *local_260 + -1;
                                                            if (*local_260 != 0) {
                                                              return plVar7;
                                                            }
                                                            lVar16 = FUN_18000cc40(local_260);
                                                            (**(code **)(lVar16 + 0x30))(plVar8);
                                                            return plVar7;
                                                          }
                                                          plVar8 = (longlong *)
                                                                   FUN_180382540(plVar5,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#52="reviews" bytes=[3953291, 3953300] */
                                                  DAT_1803b2ca0);
                                                  if (plVar8 != (longlong *)0x0) {
                                                    plVar6 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#144="editorial.md" bytes=[3954798, 3954812] */
                                                             FUN_180382540(plVar8,DAT_1803b2f80);
                                                    if ((-1 < (int)*plVar8) &&
                                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                      lVar16 = FUN_18000cc40(plVar8);
                                                      (**(code **)(lVar16 + 0x30))(plVar8);
                                                    }
                                                    if (plVar6 != (longlong *)0x0) {
                                                      plVar15 = (longlong *)FUN_1803749a0(plVar6);
                                                      if ((-1 < (int)*plVar6) &&
                                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar6);
                                                        (**(code **)(lVar16 + 0x30))(plVar6);
                                                      }
                                                      if (plVar15 != (longlong *)0x0)
                                                      goto LAB_1800d039e;
                                                    }
                                                  }
                                                  }
                                                  FUN_18000ce30(param_1,&local_2d8);
                                                  uVar17 = 0x13b;
                                                  goto LAB_1800cfec0;
                                                  }
                                                  }
                                                  }
                                                  FUN_18000ce30(param_1,&local_2d8);
                                                  uVar17 = 0x13a;
                                                  goto LAB_1800cfec0;
                                                  }
                                                  }
                                                  }
                                                  FUN_18000ce30(param_1,&local_2d8);
                                                  uVar17 = 0x139;
                                                  goto LAB_1800cfec0;
                                                  }
                                                  }
                                                  }
                                                  FUN_18000ce30(param_1,&local_2d8);
                                                  uVar17 = 0x138;
                                                  }
                                                  }
                                                  }
LAB_1800cfec0:
                                                  plVar15 = local_2b0;
                                                  plVar6 = local_2d0;
                                                  plVar8 = local_2b8;
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar7);
                                                    (**(code **)(lVar16 + 0x30))(plVar7);
                                                    plVar15 = local_2b0;
                                                    plVar6 = local_2d0;
                                                    plVar8 = local_2b8;
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1800cfeec;
                                                  }
                                                  }
                                                  }
                                                  }
                                                  FUN_18000ce30(param_1,&local_2d8);
                                                  if ((-1 < (int)*plVar5) &&
                                                     (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar5);
                                                    (**(code **)(lVar16 + 0x30))(plVar5);
                                                  }
                                                  if ((-1 < (int)*plVar15) &&
                                                     (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar15);
                                                    (**(code **)(lVar16 + 0x30))(plVar15);
                                                  }
                                                  plVar15 = local_2b0;
                                                  plVar5 = local_2c8;
                                                  plVar6 = local_2d0;
                                                  plVar8 = local_2b8;
                                                  if ((-1 < (int)*plVar4) &&
                                                     (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar4);
                                                    (**(code **)(lVar16 + 0x30))(plVar4);
                                                    plVar15 = local_2b0;
                                                    plVar5 = local_2c8;
                                                    plVar6 = local_2d0;
                                                    plVar8 = local_2b8;
                                                  }
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1800cfeec;
                                                  }
                                                  }
                                                  }
                                                  else {
                                                    uVar17 = 0x11b;
                                                    *(undefined4 *)(plVar4 + 5) = 0x11b;
                                                    plVar5 = (longlong *)
                                                             FUN_18037e980(param_1,local_2b8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#124="_stage" bytes=[3954433, 3954441] */
                                                                           DAT_1803b2ee0,
                                                                           DAT_1803b2ee8 + 0x18);
                                                    if (plVar5 != (longlong *)0x0) {
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar5);
                                                        (**(code **)(lVar16 + 0x30))(plVar5);
                                                      }
                                                      local_130 = local_240;
                                                      uVar17 = 0x11c;
                                                      *(undefined4 *)(plVar4 + 5) = 0x11c;
                                                      local_128 = local_2c8;
                                                      plVar5 = (longlong *)
                                                               FUN_18037e820(param_1,plVar8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#126="_run_editorial_review" bytes=[3954467, 3954490] */
                                                                             DAT_1803b2ef0,
                                                                             &local_130);
                                                      if (plVar5 != (longlong *)0x0) {
                                                        uVar17 = 0x11d;
                                                        *(undefined4 *)(plVar4 + 5) = 0x11d;
                                                        local_238 = plVar5;
                                                        plVar5 = (longlong *)
                                                                 FUN_18037e820(param_1,plVar8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#124="_stage" bytes=[3954433, 3954441] */
                                                                               DAT_1803b2ee0,
                                                                               DAT_1803b2ef8 + 0x18)
                                                        ;
                                                        if (plVar5 != (longlong *)0x0) {
                                                          if ((-1 < (int)*plVar5) &&
                                                             (*plVar5 = *plVar5 + -1, *plVar5 == 0))
                                                          {
                                                            lVar16 = FUN_18000cc40(plVar5);
                                                            (**(code **)(lVar16 + 0x30))(plVar5);
                                                          }
                                                          goto LAB_1800cf447;
                                                        }
                                                      }
                                                    }
                                                  }
                                                  goto LAB_1800cf062;
                                                  }
                                                  FUN_18000ce30(param_1,&local_2d8);
                                                  uVar17 = 0x11a;
                                                  plVar5 = local_2c8;
                                                  plVar6 = local_2d0;
                                                  plVar8 = local_2b8;
                                                  }
                                                  else {
                                                    lVar16 = FUN_1800c7020(param_1);
                                                    if (lVar16 == 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#4="FileRepository" bytes=[3952488, 3952504] */
                                                      FUN_180373340(param_1,&local_2d8,DAT_1803b2b20
                                                                   );
                                                      uVar17 = 0x118;
                                                      plVar5 = local_2c8;
                                                      plVar6 = local_2d0;
                                                      plVar8 = local_2b8;
                                                    }
                                                    else {
                                                      plVar14 = (longlong *)
                                                                FUN_18037f0c0(param_1,lVar16,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#61="read_text" bytes=[3953377, 3953388] */
                                                                              DAT_1803b2ce8);
                                                      if (plVar14 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_2d8);
                                                        uVar17 = 0x118;
                                                        plVar5 = local_2c8;
                                                        plVar6 = local_2d0;
                                                        plVar8 = local_2b8;
                                                      }
                                                      else {
                                                        lVar16 = FUN_1800c7240(param_1);
                                                        if (lVar16 == 0) {
                                                          FUN_180373340(param_1,&local_2d8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#123="Path" bytes=[3954427, 3954433] */
                                                                        DAT_1803b2ed8);
                                                          if ((-1 < (int)*plVar14) &&
                                                             (*plVar14 = *plVar14 + -1,
                                                             *plVar14 == 0)) {
                                                            lVar16 = FUN_18000cc40(plVar14);
                                                            (**(code **)(lVar16 + 0x30))(plVar14);
                                                          }
                                                          uVar17 = 0x118;
                                                          plVar5 = local_2c8;
                                                          plVar6 = local_2d0;
                                                          plVar8 = local_2b8;
                                                        }
                                                        else {
                                                          uVar17 = 0x118;
                                                          *(undefined4 *)(plVar4 + 5) = 0x118;
                                                          plVar7 = (longlong *)
                                                                   FUN_180376770(param_1,lVar16,
                                                                                 plVar7);
                                                          if (plVar7 != (longlong *)0x0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                            local_138 = DAT_1803b2b10;
                                                            *(undefined4 *)(plVar4 + 5) = 0x118;
                                                            local_140 = plVar7;
                                                            plVar8 = (longlong *)
                                                                     FUN_180377750(param_1,plVar14,
                                                                                   &local_140);
                                                            if ((-1 < (int)*plVar14) &&
                                                               (*plVar14 = *plVar14 + -1,
                                                               *plVar14 == 0)) {
                                                              lVar16 = FUN_18000cc40(plVar14);
                                                              (**(code **)(lVar16 + 0x30))(plVar14);
                                                            }
                                                            if ((-1 < (int)*plVar7) &&
                                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0
                                                               )) {
                                                              lVar16 = FUN_18000cc40(plVar7);
                                                              (**(code **)(lVar16 + 0x30))(plVar7);
                                                            }
                                                            if (plVar8 == (longlong *)0x0)
                                                            goto LAB_1800ce146;
                                                            plVar4 = local_2b0;
                                                            local_240 = plVar8;
                                                            if ((-1 < (int)*plVar5) &&
                                                               (*plVar5 = *plVar5 + -1, *plVar5 == 0
                                                               )) {
                                                              lVar16 = FUN_18000cc40(plVar5);
                                                              (**(code **)(lVar16 + 0x30))(plVar5);
                                                              plVar4 = local_2b0;
                                                            }
                                                            goto LAB_1800cf2d2;
                                                          }
                                                          FUN_18000ce30(param_1,&local_2d8);
                                                          plVar5 = local_2c8;
                                                          plVar6 = local_2d0;
                                                          plVar8 = local_2b8;
                                                          if (-1 < (int)*plVar14) {
                                                            *plVar14 = *plVar14 + -1;
                                                            lVar16 = *plVar14;
                                                            goto joined_r0x0001800ce5b3;
                                                          }
                                                        }
                                                      }
                                                    }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1800cfeec;
                                                  }
                                                  FUN_18000ce30(param_1,&local_2d8);
                                                  if ((-1 < (int)*plVar15) &&
                                                     (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar15);
                                                    (**(code **)(lVar16 + 0x30))(plVar15);
                                                  }
                                                  }
                                                  }
                                                  }
                                                  plVar15 = local_2b0;
                                                  plVar5 = local_2c8;
                                                  plVar6 = local_2d0;
                                                  plVar8 = local_2b8;
                                                  if ((-1 < (int)*plVar4) &&
                                                     (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar4);
                                                    (**(code **)(lVar16 + 0x30))(plVar4);
                                                    plVar15 = local_2b0;
                                                    plVar5 = local_2c8;
                                                    plVar6 = local_2d0;
                                                    plVar8 = local_2b8;
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  else {
                                                    lVar16 = FUN_180375520(param_1,plVar7,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                                           DAT_1803b2b10);
                                                    if (lVar16 == 0) {
                                                      FUN_18000ce30(param_1,&local_2d8);
                                                      uVar17 = 0xf1;
                                                      plVar5 = local_2c8;
                                                      plVar6 = local_2d0;
                                                      plVar8 = local_2b8;
                                                    }
                                                    else {
                                                      plVar14 = (longlong *)FUN_180374a40(param_1,3)
                                                      ;
                                                      FUN_18000cc80(plVar14,0,lVar16);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#105="\n\n补充信息：" bytes=[3954044, 3954063] */
                                                      if (*DAT_1803b2e48 + 1 != 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#105="\n\n补充信息：" bytes=[3954044, 3954063] */
                                                        *DAT_1803b2e48 = *DAT_1803b2e48 + 1;
                                                      }
                                                      FUN_18000cc80(plVar14,1);
                                                      lVar16 = FUN_180375520(param_1,plVar8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                                             DAT_1803b2b10);
                                                      if (lVar16 == 0) {
                                                        FUN_18000ce30(param_1,&local_2d8);
                                                        uVar17 = 0xf1;
                                                        goto LAB_1800ce5a6;
                                                      }
                                                      FUN_18000cc80(plVar14,2,lVar16);
                                                      plVar8 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                               PyUnicode_Join(DAT_1803b2b10,plVar14)
                                                      ;
                                                      if ((-1 < (int)*plVar14) &&
                                                         (*plVar14 = *plVar14 + -1, *plVar14 == 0))
                                                      {
                                                        lVar16 = FUN_18000cc40(plVar14);
                                                        (**(code **)(lVar16 + 0x30))(plVar14);
                                                      }
                                                      if (plVar8 != (longlong *)0x0) {
                                                        local_248 = plVar8;
                                                        if ((-1 < (int)*plVar7) &&
                                                           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                          lVar16 = FUN_18000cc40(plVar7);
                                                          (**(code **)(lVar16 + 0x30))(plVar7);
                                                        }
                                                        goto LAB_1800ce7a8;
                                                      }
                                                      FUN_18000ce30(param_1,&local_2d8);
                                                      uVar17 = 0xf1;
                                                      plVar5 = local_2c8;
                                                      plVar6 = local_2d0;
                                                      plVar8 = local_2b8;
                                                    }
                                                  }
                                                  }
                                                  else {
                                                    plVar14 = (longlong *)FUN_180374a40(param_1,4);
                                                    if (*piVar1 + 1 != 0) {
                                                      *piVar1 = *piVar1 + 1;
                                                    }
                                                    FUN_18000cc80(plVar14,0,piVar1);
                                                    lVar16 = FUN_180375520(param_1,plVar8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                                           DAT_1803b2b10);
                                                    if (lVar16 != 0) {
                                                      FUN_18000cc80(plVar14,1,lVar16);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#104="] " bytes=[3954040, 3954044] */
                                                      if (*DAT_1803b2e40 + 1 != 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#104="] " bytes=[3954040, 3954044] */
                                                        *DAT_1803b2e40 = *DAT_1803b2e40 + 1;
                                                      }
                                                      FUN_18000cc80(plVar14,2);
                                                      lVar16 = FUN_180375520(param_1,plVar5,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                                             DAT_1803b2b10);
                                                      if (lVar16 != 0) {
                                                        FUN_18000cc80(plVar14,3,lVar16);
                                                        plVar7 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                                 PyUnicode_Join(DAT_1803b2b10,
                                                                                plVar14);
                                                        if ((-1 < (int)*plVar14) &&
                                                           (*plVar14 = *plVar14 + -1, *plVar14 == 0)
                                                           ) {
                                                          lVar16 = FUN_18000cc40(plVar14);
                                                          (**(code **)(lVar16 + 0x30))(plVar14);
                                                        }
                                                        if (plVar7 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_2d8);
                                                          uVar17 = 0xef;
                                                          plVar5 = local_2c8;
                                                          plVar6 = local_2d0;
                                                          plVar8 = local_2b8;
                                                          goto LAB_1800cfeec;
                                                        }
                                                        plVar6 = local_2d0;
                                                        local_248 = plVar7;
                                                        if ((-1 < (int)*plVar5) &&
                                                           (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                          lVar16 = FUN_18000cc40(plVar5);
                                                          (**(code **)(lVar16 + 0x30))(plVar5);
                                                          plVar6 = local_2d0;
                                                        }
                                                        goto LAB_1800ce658;
                                                      }
                                                    }
                                                    uVar17 = 0xef;
                                                    FUN_18000ce30(param_1,&local_2d8);
LAB_1800ce5a6:
                                                    plVar15 = local_2b0;
                                                    plVar5 = local_2c8;
                                                    plVar6 = local_2d0;
                                                    plVar8 = local_2b8;
                                                    if (-1 < (int)*plVar14) {
                                                      *plVar14 = *plVar14 + -1;
                                                      lVar16 = *plVar14;
joined_r0x0001800ce5b3:
                                                      plVar5 = local_2c8;
                                                      plVar6 = local_2d0;
                                                      plVar8 = local_2b8;
                                                      if (lVar16 == 0) {
                                                        lVar16 = FUN_18000cc40(plVar14);
                                                        (**(code **)(lVar16 + 0x30))(plVar14);
                                                        plVar15 = plVar4;
                                                        plVar5 = local_2c8;
                                                        plVar6 = local_2d0;
                                                        plVar8 = local_2b8;
                                                      }
                                                    }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#94="max" bytes=[3953977, 3953982] */
                                                    uVar12 = FUN_18000d570(DAT_1803b2df0);
                                                    uVar2 = DAT_1803b2e18;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#95="min" bytes=[3953982, 3953987] */
                                                    uVar13 = FUN_18000d570(DAT_1803b2df8);
                                                    uVar11 = DAT_1803b2e28;
                                                    plVar8 = (longlong *)
                                                             FUN_180381420(local_268,plVar8);
                                                    if (plVar8 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_2d8);
                                                      uVar17 = 0xeb;
                                                      plVar15 = local_2b0;
                                                      plVar5 = local_2c8;
                                                      plVar6 = local_2d0;
                                                      plVar8 = local_2b8;
                                                    }
                                                    else {
                                                      uVar17 = 0xeb;
                                                      local_180 = uVar11;
                                                      *(undefined4 *)(local_2b0 + 5) = 0xeb;
                                                      local_178 = plVar8;
                                                      plVar6 = (longlong *)
                                                               FUN_180377750(param_1,uVar13,
                                                                             &local_180);
                                                      if ((-1 < (int)*plVar8) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar8);
                                                        (**(code **)(lVar16 + 0x30))(plVar8);
                                                      }
                                                      if (plVar6 != (longlong *)0x0) {
                                                        local_160 = uVar2;
                                                        *(undefined4 *)(local_2b0 + 5) = 0xeb;
                                                        local_158 = plVar6;
                                                        plVar5 = (longlong *)
                                                                 FUN_180377750(param_1,uVar12,
                                                                               &local_160);
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          lVar16 = FUN_18000cc40(plVar6);
                                                          (**(code **)(lVar16 + 0x30))(plVar6);
                                                        }
                                                        plVar4 = local_2b0;
                                                        if (plVar5 != (longlong *)0x0)
                                                        goto LAB_1800ce481;
                                                        goto LAB_1800ce146;
                                                      }
                                                      FUN_18000ce30(param_1,&local_2d8);
                                                      plVar15 = local_2b0;
                                                      plVar5 = local_2c8;
                                                      plVar6 = local_2d0;
                                                      plVar8 = local_2b8;
                                                    }
                                                  }
                                                }
                                              }
                                            }
                                          }
                                        }
                                      }
                                    }
                                    else {
                                      lVar16 = FUN_1800c7020(param_1);
                                      if (lVar16 == 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#4="FileRepository" bytes=[3952488, 3952504] */
                                        FUN_180373340(param_1,&local_2d8,DAT_1803b2b20);
                                        uVar17 = 0xdf;
                                        plVar6 = local_2d0;
                                        plVar8 = local_2b8;
                                      }
                                      else {
                                        plVar6 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#87="write_text" bytes=[3953859, 3953871] */
                                                 FUN_18037f0c0(param_1,lVar16,DAT_1803b2db8);
                                        if (plVar6 == (longlong *)0x0) {
                                          FUN_18000ce30(param_1,&local_2d8);
                                          uVar17 = 0xdf;
                                          plVar6 = local_2d0;
                                          plVar8 = local_2b8;
                                        }
                                        else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#50="imitation" bytes=[3953271, 3953282] */
                                          plVar4 = (longlong *)FUN_180382540(plVar5,DAT_1803b2c90);
                                          if (plVar4 != (longlong *)0x0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#91="supplementary.md" bytes=[3953929, 3953947] */
                                            plVar7 = (longlong *)FUN_180382540(plVar4,DAT_1803b2dd8)
                                            ;
                                            if ((-1 < (int)*plVar4) &&
                                               (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                                              lVar16 = FUN_18000cc40(plVar4);
                                              (**(code **)(lVar16 + 0x30))(plVar4);
                                            }
                                            if (plVar7 != (longlong *)0x0) {
                                              uVar17 = 0xdf;
                                              *(undefined4 *)(plVar15 + 5) = 0xdf;
                                              local_280 = plVar8;
                                              local_288 = plVar7;
                                              plVar8 = (longlong *)
                                                       FUN_180377750(param_1,plVar6,&local_288);
                                              if ((-1 < (int)*plVar6) &&
                                                 (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                lVar16 = FUN_18000cc40(plVar6);
                                                (**(code **)(lVar16 + 0x30))(plVar6);
                                              }
                                              if ((-1 < (int)*plVar7) &&
                                                 (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                lVar16 = FUN_18000cc40(plVar7);
                                                (**(code **)(lVar16 + 0x30))(plVar7);
                                              }
                                              if (plVar8 != (longlong *)0x0) {
                                                if ((-1 < (int)*plVar8) &&
                                                   (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                  lVar16 = FUN_18000cc40(plVar8);
                                                  (**(code **)(lVar16 + 0x30))(plVar8);
                                                }
                                                goto LAB_1800cdef7;
                                              }
                                              FUN_18000ce30(param_1,&local_2d8);
                                              plVar6 = local_2d0;
                                              plVar8 = local_2b8;
                                              goto LAB_1800cfeec;
                                            }
                                          }
                                          FUN_18000ce30(param_1,&local_2d8);
                                          if ((-1 < (int)*plVar6) &&
                                             (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                            lVar16 = FUN_18000cc40(plVar6);
                                            (**(code **)(lVar16 + 0x30))(plVar6);
                                          }
                                          uVar17 = 0xdf;
                                          plVar6 = local_2d0;
                                          plVar8 = local_2b8;
                                        }
                                      }
                                    }
                                  }
                                  else {
                                    lVar16 = FUN_1800c7020(param_1);
                                    if (lVar16 == 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#4="FileRepository" bytes=[3952488, 3952504] */
                                      FUN_180373340(param_1,&local_2d8,DAT_1803b2b20);
                                      uVar17 = 0xdd;
                                      plVar6 = local_2d0;
                                      plVar8 = local_2b8;
                                    }
                                    else {
                                      plVar6 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#87="write_text" bytes=[3953859, 3953871] */
                                               FUN_18037f0c0(param_1,lVar16,DAT_1803b2db8);
                                      if (plVar6 == (longlong *)0x0) {
                                        FUN_18000ce30(param_1,&local_2d8);
                                        uVar17 = 0xdd;
                                        plVar6 = local_2d0;
                                        plVar8 = local_2b8;
                                      }
                                      else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#50="imitation" bytes=[3953271, 3953282] */
                                        plVar4 = (longlong *)FUN_180382540(plVar5,DAT_1803b2c90);
                                        if (plVar4 != (longlong *)0x0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#90="inspiration_tags.md" bytes=[3953908, 3953929] */
                                          plVar7 = (longlong *)FUN_180382540(plVar4,DAT_1803b2dd0);
                                          if ((-1 < (int)*plVar4) &&
                                             (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                                            lVar16 = FUN_18000cc40(plVar4);
                                            (**(code **)(lVar16 + 0x30))(plVar4);
                                          }
                                          if (plVar7 != (longlong *)0x0) {
                                            uVar17 = 0xdd;
                                            *(undefined4 *)(plVar15 + 5) = 0xdd;
                                            local_280 = plVar8;
                                            local_288 = plVar7;
                                            plVar8 = (longlong *)
                                                     FUN_180377750(param_1,plVar6,&local_288);
                                            if ((-1 < (int)*plVar6) &&
                                               (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                              lVar16 = FUN_18000cc40(plVar6);
                                              (**(code **)(lVar16 + 0x30))(plVar6);
                                            }
                                            if ((-1 < (int)*plVar7) &&
                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                              lVar16 = FUN_18000cc40(plVar7);
                                              (**(code **)(lVar16 + 0x30))(plVar7);
                                            }
                                            if (plVar8 != (longlong *)0x0) {
                                              if ((-1 < (int)*plVar8) &&
                                                 (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                lVar16 = FUN_18000cc40(plVar8);
                                                (**(code **)(lVar16 + 0x30))(plVar8);
                                              }
                                              goto LAB_1800cdd65;
                                            }
                                            FUN_18000ce30(param_1,&local_2d8);
                                            plVar6 = local_2d0;
                                            plVar8 = local_2b8;
                                            goto LAB_1800cfeec;
                                          }
                                        }
                                        FUN_18000ce30(param_1,&local_2d8);
                                        if ((-1 < (int)*plVar6) &&
                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                          lVar16 = FUN_18000cc40(plVar6);
                                          (**(code **)(lVar16 + 0x30))(plVar6);
                                        }
                                        uVar17 = 0xdd;
                                        plVar6 = local_2d0;
                                        plVar8 = local_2b8;
                                      }
                                    }
                                  }
                                }
                                else {
                                  lVar16 = FUN_1800c7020(param_1);
                                  if (lVar16 == 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#4="FileRepository" bytes=[3952488, 3952504] */
                                    FUN_180373340(param_1,&local_2d8,DAT_1803b2b20);
                                    uVar17 = 0xdb;
                                    plVar6 = local_2d0;
                                    plVar8 = local_2b8;
                                  }
                                  else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#87="write_text" bytes=[3953859, 3953871] */
                                    plVar6 = (longlong *)FUN_18037f0c0(param_1,lVar16,DAT_1803b2db8)
                                    ;
                                    if (plVar6 == (longlong *)0x0) {
                                      FUN_18000ce30(param_1,&local_2d8);
                                      uVar17 = 0xdb;
                                      plVar6 = local_2d0;
                                      plVar8 = local_2b8;
                                    }
                                    else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#50="imitation" bytes=[3953271, 3953282] */
                                      plVar4 = (longlong *)FUN_180382540(plVar5,DAT_1803b2c90);
                                      if (plVar4 != (longlong *)0x0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#89="reader_preferences.md" bytes=[3953885, 3953908] */
                                        plVar7 = (longlong *)FUN_180382540(plVar4,DAT_1803b2dc8);
                                        if ((-1 < (int)*plVar4) &&
                                           (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                                          lVar16 = FUN_18000cc40(plVar4);
                                          (**(code **)(lVar16 + 0x30))(plVar4);
                                        }
                                        if (plVar7 != (longlong *)0x0) {
                                          uVar17 = 0xdb;
                                          *(undefined4 *)(plVar15 + 5) = 0xdb;
                                          local_280 = plVar8;
                                          local_288 = plVar7;
                                          plVar8 = (longlong *)
                                                   FUN_180377750(param_1,plVar6,&local_288);
                                          if ((-1 < (int)*plVar6) &&
                                             (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                            lVar16 = FUN_18000cc40(plVar6);
                                            (**(code **)(lVar16 + 0x30))(plVar6);
                                          }
                                          if ((-1 < (int)*plVar7) &&
                                             (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                            lVar16 = FUN_18000cc40(plVar7);
                                            (**(code **)(lVar16 + 0x30))(plVar7);
                                          }
                                          if (plVar8 != (longlong *)0x0) {
                                            if ((-1 < (int)*plVar8) &&
                                               (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                              lVar16 = FUN_18000cc40(plVar8);
                                              (**(code **)(lVar16 + 0x30))(plVar8);
                                            }
                                            goto LAB_1800cdbd3;
                                          }
                                          FUN_18000ce30(param_1,&local_2d8);
                                          plVar6 = local_2d0;
                                          plVar8 = local_2b8;
                                          goto LAB_1800cfeec;
                                        }
                                      }
                                      FUN_18000ce30(param_1,&local_2d8);
                                      if ((-1 < (int)*plVar6) &&
                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                        lVar16 = FUN_18000cc40(plVar6);
                                        (**(code **)(lVar16 + 0x30))(plVar6);
                                      }
                                      uVar17 = 0xdb;
                                      plVar6 = local_2d0;
                                      plVar8 = local_2b8;
                                    }
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                      goto LAB_1800cfeec;
                    }
                  }
                  local_2d8 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                }
              }
            }
          }
        }
      }
    }
  }
LAB_1800cfeec:
  piVar1 = *(int **)(local_2d8 + 0x28);
  if (piVar1 == (int *)0x0) {
    lVar16 = FUN_18038e2a0(plVar15,uVar17);
  }
  else {
    if (*(longlong **)(piVar1 + 6) == plVar15) goto LAB_1800cff2e;
    lVar16 = FUN_18038e2a0(plVar15,uVar17);
    *(int **)(lVar16 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
  }
  FUN_180001300(&local_2d8,lVar16);
LAB_1800cff2e:
  plVar7 = local_248;
  local_2e0 = local_238;
  local_2e8 = local_240;
  local_2f0 = local_1f0;
  local_2f8 = local_1f8;
  local_300 = local_1a8;
  local_308 = local_1b0;
  local_310 = local_1b8;
  local_318 = local_1c0;
  local_320 = local_248;
  local_328 = local_250;
  local_330 = local_258;
  local_338 = local_218;
  local_340 = local_1c8;
  local_350 = local_1d8;
  local_358 = local_1e0;
  local_360 = local_260;
  local_370 = local_278;
  local_378 = local_220;
  local_380 = local_1e8;
  local_388 = local_210;
  local_390 = local_228;
  local_398 = local_268;
  local_368 = plVar5;
  local_348 = plVar6;
  FUN_18036d0a0(plVar15,"oooooooooooooooooooooooooo",plVar8,local_230);
  plVar4 = DAT_1803b29f0;
  if (plVar15 == DAT_1803b29f0) {
    if ((-1 < (int)*DAT_1803b29f0) && (*DAT_1803b29f0 = *DAT_1803b29f0 + -1, *plVar4 == 0)) {
      (**(code **)(plVar4[1] + 0x30))();
    }
    DAT_1803b29f0 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  lVar16 = local_2d8;
  local_2d8 = 0;
  if (((local_1e0 != (longlong *)0x0) && (-1 < (int)*local_1e0)) &&
     (*local_1e0 = *local_1e0 + -1, *local_1e0 == 0)) {
    (**(code **)(local_1e0[1] + 0x30))();
  }
  if (((local_1d8 != (longlong *)0x0) && (-1 < (int)*local_1d8)) &&
     (*local_1d8 = *local_1d8 + -1, *local_1d8 == 0)) {
    (**(code **)(local_1d8[1] + 0x30))();
  }
  if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)
     ) {
    (**(code **)(plVar6[1] + 0x30))(plVar6);
  }
  if (((local_1c8 != (longlong *)0x0) && (-1 < (int)*local_1c8)) &&
     (*local_1c8 = *local_1c8 + -1, *local_1c8 == 0)) {
    (**(code **)(local_1c8[1] + 0x30))();
  }
  if (((local_218 != (code *)0x0) && (-1 < *(int *)local_218)) &&
     (*(longlong *)local_218 = *(longlong *)local_218 + -1, *(longlong *)local_218 == 0)) {
    (**(code **)(*(longlong *)(local_218 + 8) + 0x30))();
  }
  if (((local_258 != (longlong *)0x0) && (-1 < (int)*local_258)) &&
     (*local_258 = *local_258 + -1, *local_258 == 0)) {
    (**(code **)(local_258[1] + 0x30))();
  }
  if (((local_250 != (longlong *)0x0) && (-1 < (int)*local_250)) &&
     (*local_250 = *local_250 + -1, *local_250 == 0)) {
    (**(code **)(local_250[1] + 0x30))();
  }
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  if (((local_1c0 != (longlong *)0x0) && (-1 < (int)*local_1c0)) &&
     (*local_1c0 = *local_1c0 + -1, *local_1c0 == 0)) {
    (**(code **)(local_1c0[1] + 0x30))();
  }
  if (((local_1b8 != (longlong *)0x0) && (-1 < (int)*local_1b8)) &&
     (*local_1b8 = *local_1b8 + -1, *local_1b8 == 0)) {
    (**(code **)(local_1b8[1] + 0x30))();
  }
  if (((local_1b0 != (longlong *)0x0) && (-1 < (int)*local_1b0)) &&
     (*local_1b0 = *local_1b0 + -1, *local_1b0 == 0)) {
    (**(code **)(local_1b0[1] + 0x30))();
  }
  if (((local_1a8 != (longlong *)0x0) && (-1 < (int)*local_1a8)) &&
     (*local_1a8 = *local_1a8 + -1, *local_1a8 == 0)) {
    (**(code **)(local_1a8[1] + 0x30))();
  }
  if (((local_1f8 != (longlong *)0x0) && (-1 < (int)*local_1f8)) &&
     (*local_1f8 = *local_1f8 + -1, *local_1f8 == 0)) {
    (**(code **)(local_1f8[1] + 0x30))();
  }
  if (((local_1f0 != (longlong *)0x0) && (-1 < (int)*local_1f0)) &&
     (*local_1f0 = *local_1f0 + -1, *local_1f0 == 0)) {
    (**(code **)(local_1f0[1] + 0x30))();
  }
  if (((local_240 != (longlong *)0x0) && (-1 < (int)*local_240)) &&
     (*local_240 = *local_240 + -1, *local_240 == 0)) {
    (**(code **)(local_240[1] + 0x30))();
  }
  if (((local_238 != (longlong *)0x0) && (-1 < (int)*local_238)) &&
     (*local_238 = *local_238 + -1, *local_238 == 0)) {
    (**(code **)(local_238[1] + 0x30))();
  }
  local_2d8 = lVar16;
  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
    (**(code **)(plVar8[1] + 0x30))(plVar8);
  }
  if ((-1 < (int)*local_230) && (*local_230 = *local_230 + -1, *local_230 == 0)) {
    (**(code **)(local_230[1] + 0x30))();
  }
  if ((-1 < (int)*local_268) && (*local_268 = *local_268 + -1, *local_268 == 0)) {
    (**(code **)(local_268[1] + 0x30))();
  }
  if ((-1 < (int)*local_228) && (*local_228 = *local_228 + -1, *local_228 == 0)) {
    (**(code **)(local_228[1] + 0x30))();
  }
  if ((-1 < (int)*local_210) && (*local_210 = *local_210 + -1, *local_210 == 0)) {
    (**(code **)(local_210[1] + 0x30))();
  }
  if ((-1 < (int)*local_1e8) && (*local_1e8 = *local_1e8 + -1, *local_1e8 == 0)) {
    (**(code **)(local_1e8[1] + 0x30))();
  }
  if ((-1 < (int)*local_220) && (*local_220 = *local_220 + -1, *local_220 == 0)) {
    (**(code **)(local_220[1] + 0x30))();
  }
  if ((-1 < (int)*local_278) && (*local_278 = *local_278 + -1, *local_278 == 0)) {
    (**(code **)(local_278[1] + 0x30))();
  }
  if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
    (**(code **)(plVar5[1] + 0x30))(plVar5);
  }
  if ((-1 < (int)*local_260) && (*local_260 = *local_260 + -1, *local_260 == 0)) {
    (**(code **)(local_260[1] + 0x30))();
  }
  FUN_18000ce40(param_1,&local_2d8);
  return (longlong *)0x0;
}
